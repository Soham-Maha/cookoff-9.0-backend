package controllers

import (
	"context"
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Question struct {
	ID           uuid.UUID   `json:"id"`
	Description  *string     `json:"description"`
	Title        *string     `json:"title"`
	InputFormat  *string     `json:"input_format"`
	Points       pgtype.Int4 `json:"points"`
	Round        int32       `json:"round"`
	Constraints  *string     `json:"constraints"`
	OutputFormat *string     `json:"output_format"`
}

func GetAllQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fetchedQuestions, err := database.Queries.GetQuestions(ctx)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httphelpers.WriteJSON(w, 200, fetchedQuestions)
}

func GetQuestionById(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	questionIdStr := chi.URLParam(r, "question_id")
	question_id, err := uuid.Parse(questionIdStr)

	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	fetchedQuestion, err := database.Queries.GetQuestion(ctx, question_id)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, err)
		return
	}

	httphelpers.WriteJSON(w, 200, fetchedQuestion)
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var question Question
	err := httphelpers.ParseJSON(r, &question)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	questions, err := database.Queries.CreateQuestion(ctx, db.CreateQuestionParams{
		ID:           uuid.New(),
		Description:  question.Description,
		Title:        question.Title,
		InputFormat:  question.InputFormat,
		Points:       question.Points,
		Round:        question.Round,
		Constraints:  question.Constraints,
		OutputFormat: question.OutputFormat,
	})
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httphelpers.WriteJSON(w, 200, questions)
}

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	questionIdStr := chi.URLParam(r, "question_id")
	question_id, err := uuid.Parse(questionIdStr)

	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = database.Queries.DeleteQuestion(ctx, question_id)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, err)
		return
	}
}
