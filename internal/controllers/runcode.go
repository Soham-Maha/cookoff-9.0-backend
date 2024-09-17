package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/google/uuid"
)

type judgeresp struct {
	TestCaseID     string
	StdOut         string `json:"stdout"`
	Time           string `json:"time"`
	Memory         int    `json:"memory"`
	StdErr         string `json:"stderr"`
	Token          string `json:"token"`
	Message        string `json:"message"`
	Status         Status `json:"status"`
	CompilerOutput string `json:"compile_output"`
}

type subpayload struct {
	LanguageID int    `json:"language_id"`
	SourceCode string `json:"source_code"`
	Input      string `json:"stdin"`
	Output     string `json:"expected_output"`
}

type resp struct {
	Result []judgeresp `json:"result"`
}

func RunCode(w http.ResponseWriter, r *http.Request) {
	JUDGE0_URI := os.Getenv("JUDGE0_URI")
	ctx := r.Context()

	var req subreq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	question_id, _ := uuid.Parse(req.QuestionID)

	testcases, err := database.Queries.GetTestCases(ctx, db.GetTestCasesParams{QuestionID: question_id, Column2: true})
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Question not found")
		return
	}

	judge0URL := JUDGE0_URI + "/submissions/?base64_encoded=true&wait=true"

	var payload subpayload
	response := resp{
		Result: make([]judgeresp, len(testcases)),
	}
	for i, testcase := range testcases {
		payload = subpayload{
			LanguageID: req.LanguageID,
			SourceCode: b64(req.SourceCode),
			Input:      b64(*testcase.Input),
			Output:     b64(*testcase.ExpectedOutput),
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			logger.Errof("Error marshaling payload: %v", err)
			httphelpers.WriteError(w, http.StatusInternalServerError, "Unable to marshal payload")
			return
		}

		logger.Infof(string(payloadJSON))

		result, err := http.Post(judge0URL, "application/json", bytes.NewBuffer(payloadJSON))
		if err != nil {
			logger.Errof("Error making request to Judge0: %v", err)
			httphelpers.WriteError(w, http.StatusInternalServerError, "Error making request to Judge0")
			return
		}
		defer result.Body.Close()

		var data judgeresp
		data.TestCaseID = testcase.ID.String()
		if err = json.NewDecoder(result.Body).Decode(&data); err != nil {
			logger.Errof("Error decoding response from Judge0: %v", err)
			httphelpers.WriteError(w, http.StatusInternalServerError, "Error decoding response from Judge0")
			return
		}

		data.CompilerOutput, _ = db64(data.CompilerOutput)
		response.Result[i] = data
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errof("Error encoding response: %v", err)
		httphelpers.WriteError(w, http.StatusInternalServerError, "Error encoding response")
	}
}

func b64(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func db64(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}
