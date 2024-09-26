package worker

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"strconv"

	"github.com/CodeChefVIT/cookoff-backend/internal/controllers"
	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/submission"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
)

const TypeProcessSubmission = "submission:process"
const SubmissionDoneStatus = "DONE"

func calculateScore() float64 {
	return 7.4
}

// ProcessSubmissionTask processes the submission task based on status.
func ProcessSubmissionTask(ctx context.Context, t *asynq.Task) error {
	var data controllers.Data
	logger.Infof("Processing task: %v", t.Type)
	logger.Infof("Payload: %v", string(t.Payload()))

	if err := json.Unmarshal(t.Payload(), &data); err != nil {
		log.Printf("Error unmarshalling task payload: %v\n", err)
		return err
	}

	timeValue, err := parseTime(data.Time)
	if err != nil {
		log.Println("Error parsing time value: ", err)
		return err
	}

	value, err := submission.GetSubID(ctx, data.Token)
	if err != nil {
		log.Println("Error getting submission ID from token: ", err)
		return err
	}

	idUUID, err := uuid.Parse(value)
	if err != nil {
		log.Fatalf("Error parsing UUID: %v", err)
	}

	sub, err := database.Queries.GetSubmission(ctx, idUUID)
	if err != nil {
		log.Println("Error retrieving submission: ", err)
		return err
	}

	testcasesPassed := int(sub.TestcasesPassed.Int32)
	testcasesFailed := int(sub.TestcasesFailed.Int32)

	switch data.Status.ID {
	case "3":
		testcasesPassed++
	case "4":
		testcasesFailed++
	case "6":
		err = handleCompilationError(ctx, idUUID, data)
	case "11":
		err = handleRuntimeError(ctx, idUUID)
	}

	if err != nil {
		return err
	}

	err = updateSubmission(ctx, idUUID, testcasesPassed, testcasesFailed, timeValue, data.Memory)
	if err != nil {
		return err
	}

	if err := submission.Tokens.DeleteToken(ctx, data.Token); err != nil {
		log.Println("Error deleting token: ", err)
		return err
	}

	tokenCount, err := submission.Tokens.GetTokenCount(ctx, value)
	if err != nil {
		log.Println("Error getting token count: ", err)
		return err
	}

	if tokenCount == 0 {
		err = finalizeSubmission(ctx, idUUID)
		if err != nil {
			return err
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

func handleCompilationError(ctx context.Context, idUUID uuid.UUID, data controllers.Data) error {
	subID, err := uuid.NewV7()
	log.Printf("Compilation error for submission %v\n", idUUID)
	err = database.Queries.UpdateSubmission(ctx, db.UpdateSubmissionParams{
		TestcasesPassed: pgtype.Int4{Int32: int32(0), Valid: true},
		TestcasesFailed: pgtype.Int4{Int32: int32(0), Valid: true},
		Runtime:         pgtype.Numeric{Int: big.NewInt(0), Valid: true},
		Memory:          pgtype.Numeric{Int: big.NewInt(0), Valid: true},
		ID:              idUUID,
	})

	if err != nil {
		log.Println("Error updating submission for compilation error: ", err)
		return err
	}

	err = database.Queries.CreateSubmissionStatus(ctx, db.CreateSubmissionStatusParams{
		ID:           subID,
		SubmissionID: idUUID,
		Runtime:      pgtype.Numeric{Int: big.NewInt(0), Valid: true},
		Memory:       pgtype.Numeric{Int: big.NewInt(0), Valid: true},
		Description:  &data.Status.Description,
	})
	return nil
}

func handleRuntimeError(ctx context.Context, idUUID uuid.UUID) error {
	log.Printf("Runtime error for submission %v\n", idUUID)
	err := database.Queries.UpdateSubmission(ctx, db.UpdateSubmissionParams{
		TestcasesPassed: pgtype.Int4{Int32: int32(0), Valid: true},
		TestcasesFailed: pgtype.Int4{Int32: int32(0), Valid: true},
		Runtime:         pgtype.Numeric{Int: big.NewInt(0), Valid: true},
		Memory:          pgtype.Numeric{Int: big.NewInt(0), Valid: true},
		ID:              idUUID,
	})

	if err != nil {
		log.Println("Error updating submission for runtime error: ", err)
		return err
	}

	notAcceptedStatus := "NOT ACCEPTED"
	err = database.Queries.UpdateDescriptionStatus(ctx, db.UpdateDescriptionStatusParams{
		Description: &notAcceptedStatus,
		ID:          idUUID,
	})

	if err != nil {
		log.Println("Error updating submission status to 'Not Accepted': ", err)
		return err
	}
	return nil
}

func updateSubmission(ctx context.Context, idUUID uuid.UUID, testcasesPassed, testcasesFailed int, timeValue float64, memory int) error {
	err := database.Queries.UpdateSubmission(ctx, db.UpdateSubmissionParams{
		TestcasesPassed: pgtype.Int4{Int32: int32(testcasesPassed), Valid: true},
		TestcasesFailed: pgtype.Int4{Int32: int32(testcasesFailed), Valid: true},
		Runtime:         pgtype.Numeric{Int: big.NewInt(int64(timeValue * 1000)), Valid: true},
		Memory:          pgtype.Numeric{Int: big.NewInt(int64(memory)), Valid: true},
		ID:              idUUID,
	})

	if err != nil {
		log.Println("Error updating submission: ", err)
		return err
	}

	log.Printf("Submission ID: %v Testcases Passed: %v Testcases Failed: %v\n", idUUID, testcasesPassed, testcasesFailed)
	return nil
}

func finalizeSubmission(ctx context.Context, idUUID uuid.UUID) error {
	status := SubmissionDoneStatus
	err := database.Queries.UpdateSubmissionStatus(ctx, db.UpdateSubmissionStatusParams{
		Status: &status,
		ID:     idUUID,
	})

	if err != nil {
		log.Println("Error updating submission status to done: ", err)
		return err
	}
	return nil
}
