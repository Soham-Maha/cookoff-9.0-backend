package controllers

import (
	"context"
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func CreateTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	var testCase db.CreateTestCaseParams
	if err := httphelpers.ParseJSON(r, &testCase); err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	testCase.ID = uuid.New()

	err := database.Queries.CreateTestCase(context.Background(), testCase)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "Failed to create test case")
		return
	}

	httphelpers.WriteJSON(w, http.StatusCreated, testCase)
}

func GetTestCaseHandler(w http.ResponseWriter, r *http.Request) {

	testcaseID, err := uuid.Parse(chi.URLParam(r, "testcase_id"))
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid test case ID")
		return
	}

	testCase, err := database.Queries.GetTestCase(context.Background(), testcaseID)
	if err != nil {
		httphelpers.WriteError(w, http.StatusNotFound, "Test case not found")
		return
	}

	httphelpers.WriteJSON(w, http.StatusOK, testCase)
}

func GetAllTestCasesHandler(w http.ResponseWriter, r *http.Request) {

	testCases, err := database.Queries.GetAllTestCases(context.Background())
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "Failed to fetch test cases")
		return
	}

	httphelpers.WriteJSON(w, http.StatusOK, testCases)
}

func UpdateTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	testcaseID, err := uuid.Parse(chi.URLParam(r, "testcase_id"))
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid test case ID")
		return
	}

	var updateData db.UpdateTestCaseParams

	if err := httphelpers.ParseJSON(r, &updateData); err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	updateData.ID = testcaseID

	err = database.Queries.UpdateTestCase(context.Background(), updateData)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "Failed to update test case")
		return
	}

	httphelpers.WriteJSON(w, http.StatusOK, map[string]string{"message": "Test case updated successfully"})
}

func DeleteTestCaseHandler(w http.ResponseWriter, r *http.Request) {

	testcaseID, err := uuid.Parse(chi.URLParam(r, "testcase_id"))
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid test case ID")
		return
	}

	err = database.Queries.DeleteTestCase(context.Background(), testcaseID)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "Failed to delete test case")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
