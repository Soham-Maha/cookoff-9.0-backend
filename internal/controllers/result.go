package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/submission"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/validator"

	"github.com/google/uuid"
)

type resultreq struct {
	SubID string `json:"submission_id" validate:"required"`
}

func GetResult(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*2)
	defer cancel()

	var req resultreq
	err := httphelpers.ParseJSON(r, &req)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := validator.ValidatePayload(w, req); err != nil {
		httphelpers.WriteError(w, http.StatusNotAcceptable, "Please provide values for all required fields.")
		return
	}

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

	ticker := time.NewTicker(20 * time.Second)
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

	var req string = "https://judge0-ce.p.sulu.sh/submissions/batch?tokens=" + strings.Join(members, ",")
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
		}
	}

	return nil
}
