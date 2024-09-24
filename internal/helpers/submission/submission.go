package submission

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	"github.com/google/uuid"
)

type Token struct {
	Token string `json:"token"`
}

type Payload struct {
	Submissions []Submission `json:"submissions"`
}

func CreateSubmission(ctx context.Context, question_id uuid.UUID, language_id int, source string) ([]byte, error) {
	callback_url := os.Getenv("CALLBACK_URL")

	testcases, err := database.Queries.GetTestCases(ctx, db.GetTestCasesParams{QuestionID: question_id, Column2: false})
	if err != nil {
		return nil, fmt.Errorf("error getting test cases for question_id %d: %v", question_id, err)
	}
	payload := Payload{
		Submissions: make([]Submission, len(testcases)),
	}

	runtime_mut, err := RuntimeMut(language_id)
	if err != nil {
		return nil, err
	}

	for i, testcase := range testcases {
		runtime, _ := testcase.Runtime.Float64Value()
		payload.Submissions[i] = Submission{
			SourceCode: B64(source),
			LanguageID: language_id,
			Input:      B64(testcase.Input),
			Output:     B64(testcase.ExpectedOutput),
			Runtime:    runtime.Float64 * float64(runtime_mut),
			Callback:   callback_url,
		}
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %v", err)
	}

	return payloadJSON, nil
}

func StoreTokens(ctx context.Context, subID uuid.UUID, r *http.Response) error {
	var tokens []Token
	err := json.NewDecoder(r.Body).Decode(&tokens)
	if err != nil {
		return fmt.Errorf("Invalid request payload")
	}

	for _, t := range tokens {
		err := Tokens.AddToken(ctx, t.Token, subID.String())
		if err != nil {
			return fmt.Errorf("Failed to add token")
		}
	}
	return nil
}

func GetSubID(ctx context.Context, token string) (string, error) {
	subID, err := Tokens.GetSubID(ctx, token)
	if err != nil {
		return "", err
	}
	return subID, nil
}
