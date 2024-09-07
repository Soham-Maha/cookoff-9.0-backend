package submission

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/google/uuid"
)

type Submission struct {
	LanguageID int    `json:"language_id"`
	SourceCode string `json:"source_code"`
	Input      string `json:"stdin"`
	Output     string `json:"expected_output"`
	Callback   string `json:"callback_url"`
}

type Payload struct {
	Submissions []Submission `json:"submissions"`
}

func CreateSubmission(ctx context.Context, questionID string, language_id int, source string) ([]byte, error) {
	callback_url := os.Getenv("CALLBACK_URL")

	question_id, err := uuid.Parse(questionID)
	if err != nil {
		return nil, err
	}
	query := db.New(database.DBPool)
	testcases, err := query.GetTestCases(ctx, question_id)
	if err != nil {
		logger.Errof("Error getting test cases for question_id %d: %v", question_id, err)
		return nil, err
	}
	payload := Payload{
		Submissions: make([]Submission, len(testcases)),
	}

	for i, testcase := range testcases {
		payload.Submissions[i] = Submission{
			SourceCode: b64(source),
			LanguageID: language_id,
			Input:      b64(testcase.Input.String),
			Output:     b64(testcase.ExpectedOutput.String),
			Callback:   callback_url,
		}
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logger.Errof("Error marshaling payload: %v", err)
		return nil, err
	}

	return payloadJSON, nil
}

func b64(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}
