package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/submission"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"
)

type resultreq struct {
	SubID string `json:"submission_id" validate:"required"`
}

func GetResult(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*2)
	defer cancel()

	var req resultreq

	req.SubID = chi.URLParam(r, "submission_id")
	subid, err := uuid.Parse(req.SubID)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	processed, err := submission.CheckStatus(ctx, subid)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error while getting status")
		return
	}

	if processed {
		result, err := submission.GetSubResult(ctx, subid)
		if err != nil {
			httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error while getting submission result")
			return
		}
		httphelpers.WriteJSON(w, http.StatusOK, result)
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			httphelpers.WriteError(w, http.StatusRequestTimeout, "Submission not processed")
			return

		case <-ticker.C:
			err := BadCodeAlert(ctx, subid, w)
			if err != nil {
				return
			}
			processed, err := submission.CheckStatus(ctx, subid)
			if err != nil {
				httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error while getting status")
				return
			}

			if processed {
				result, err := submission.GetSubResult(ctx, subid)
				if err != nil {
					httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error while getting submission result")
					return
				}
				httphelpers.WriteJSON(w, http.StatusOK, result)
				return
			}
		}
	}

}

type GetStatus struct {
	Description string `json:"description"`
	ID          int    `json:"id"`
}

type GetSub struct {
	CompileOutput *string   `json:"compile_output"`
	Memory        *int      `json:"memory"`
	Message       *string   `json:"message"`
	Status        GetStatus `json:"status"`
	Stderr        *string   `json:"stderr"`
	Stdout        *string   `json:"stdout"`
	Time          *string   `json:"time"`
	Token         string    `json:"token"`
}

type Response struct {
	Submissions []GetSub `json:"submissions"`
}

func BadCodeAlert(ctx context.Context, id uuid.UUID, w http.ResponseWriter) error {

	members, err := submission.Tokens.GetTokenMember(ctx, id.String())
	if err != nil {
		return err
	}

	if len(members) == 0 {
		err := submission.UpdateSubmission(ctx, id)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error!(Failed to update submission table)")
			return err
		}
		return nil
	}

	var req string = "https://judge0-ce.p.sulu.sh/submissions/batch?base64_encoded=true&tokens=" + strings.Join(members, ",")
	fmt.Println("urk :- ", req)
	resp, err := submission.BatchGet(req)
	if err != nil {
		logger.Errof("Error sending request to Judge0: %v", err)
		httphelpers.WriteError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("Error sending request to Judge0: %v", err),
		)
		return err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errof("Error reading response body from Judge0: %v", err)
		httphelpers.WriteError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("Error reading response body from Judge0: %v", err),
		)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Errof(
			"Unexpected status code from Judge0: %d, error: %v",
			resp.StatusCode,
			string(respBytes),
		)
		httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error!")
		return err
	}

	var temp Response
	err = json.Unmarshal(respBytes, &temp)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error!(Failed to unmarshal the response)")
		return err
	}

	if count, err := submission.Tokens.GetTokenCount(ctx, id.String()); count == 0 || err != nil {
		err := submission.UpdateSubmission(ctx, id)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error!(Failed to update submission table)")
			return err
		}
	}

	for _, v := range temp.Submissions {
		if v.Status.ID != 1 || v.Status.ID != 2 {
			submission.Tokens.DeleteToken(ctx, v.Token)
			_, testcase, err := submission.GetSubID(ctx, v.Token)
			if err != nil {
				fmt.Println("Failed to get details from redis:", err)
				return nil
				//httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error!(Failed to get from redis)")
				//return err
			}
			timeValue, err := parseTime(*v.Time)
			if err != nil {
				fmt.Println("Error converting to float:", err)
				httphelpers.WriteError(w, http.StatusInternalServerError, "Internal server error!(Failed to convert to float)")
				return err
			}
			tid := uuid.MustParse(testcase)
			switch v.Status.ID {
			case 3:
				err = HandleCompilationError(ctx, id, v, int(timeValue*1000), tid, "success")
			case 4:
				err = HandleCompilationError(ctx, id, v, int(timeValue*1000), tid, "wrong answer")
			case 6:
				err = HandleCompilationError(ctx, id, v, int(timeValue*1000), tid, "Compilation error")
			case 11:
				err = HandleCompilationError(ctx, id, v, int(timeValue*1000), tid, "Runtime error")
			}
			if err != nil {
				fmt.Println("Failed to add submission_results")
			}
		}
	}

	return nil
}

func parseTime(timeStr string) (float64, error) {
	if timeStr == "" {
		log.Println("Time value is empty, setting time to 0 for this submission.")
		return 0, nil
	}

	timeValue, err := strconv.ParseFloat(timeStr, 64)
	if err != nil {
		return 0, err
	}
	return timeValue, nil
}

func HandleCompilationError(ctx context.Context, idUUID uuid.UUID, data GetSub, time int, testcase uuid.UUID, status string) error {
	subID, err := uuid.NewV7()

	if err != nil {
		log.Println("Error updating submission for compilation error: ", err)
		return err
	}

	err = database.Queries.CreateSubmissionStatus(ctx, db.CreateSubmissionStatusParams{
		ID:           subID,
		SubmissionID: idUUID,
		TestcaseID:   uuid.NullUUID{UUID: testcase, Valid: true},
		Runtime:      pgtype.Numeric{Int: big.NewInt(int64(time)), Valid: true},
		Memory:       pgtype.Numeric{Int: big.NewInt(int64(*data.Memory)), Valid: true},
		Description:  &data.Status.Description,
		Status:       status,
	})

	if err != nil {
		log.Println("Error creating submission status error: ", err)
		return err
	}
	return nil
}
