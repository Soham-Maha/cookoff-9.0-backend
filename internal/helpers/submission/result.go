package submission

import (
	"context"
	"fmt"

	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/google/uuid"
)

func CheckStatus(ctx context.Context, subid uuid.UUID) (bool, error) {
	status, err := database.Queries.GetSubmissionStatusByID(ctx, subid)
	if err != nil {
		logger.Errof(fmt.Sprintf("Error while getting submission status: %v", err.Error()))
		return false, err
	}

	if status == nil {
		return false, nil
	}

	return *status == "DONE", nil
}

func GetSubResult(ctx context.Context, subid uuid.UUID) (resultresp, error) {
	submission, err := database.Queries.GetSubmissionByID(ctx, subid)
	if err != nil {
		logger.Errof(fmt.Sprintf("Error while getting submission result: %v", err.Error()))
		return resultresp{}, err
	}

	sub_result, err := database.Queries.GetSubmissionResultsBySubmissionID(ctx, subid)
	if err != nil {
		logger.Errof(fmt.Sprintf("Error while getting submission results: %v", err.Error()))
		return resultresp{}, err
	}

	var desc string
	if submission.Description == nil {
		desc = ""
	} else {
		desc = *submission.Description
	}

	sub_runtime, _ := submission.Runtime.Float64Value()
	sub_memory, _ := submission.Memory.Float64Value()
	resp := resultresp{
		ID:             submission.ID.String(),
		QuestionID:     submission.QuestionID.String(),
		Passed:         int(submission.TestcasesPassed.Int32),
		Failed:         int(submission.TestcasesFailed.Int32),
		Runtime:        sub_runtime.Float64,
		Memory:         sub_memory.Float64,
		SubmissionTime: submission.SubmissionTime.Time.String(),
		Description:    desc,
		Testcases:      make([]tc_result, len(sub_result)),
	}

	for i, result := range sub_result {
		runtime, _ := result.Runtime.Float64Value()
		memory, _ := submission.Memory.Float64Value()
		var testcase_id string
		if result.TestcaseID.Valid {
			testcase_id = result.TestcaseID.UUID.String()
		} else {
			testcase_id = ""
		}
		resp.Testcases[i] = tc_result{
			ID:          testcase_id,
			Runtime:     runtime.Float64,
			Memory:      memory.Float64,
			Status:      result.Status,
			Description: *result.Description,
		}
	}

	return resp, nil
}
