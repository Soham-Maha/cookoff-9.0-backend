package server

import (
	"net/http"
    custom_middleware "github.com/CodeChefVIT/cookoff-backend/internal/middlewares"
	"github.com/CodeChefVIT/cookoff-backend/internal/controllers"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	tokenManager := auth.Tokens
	r.With(custom_middleware.VerifyRefreshTokenMiddleware(tokenManager)).Post("/token/refresh",controllers.RefreshToken)
	r.Get("/ping", controllers.HealthCheck)
	r.Post("/submit", controllers.SubmitCode)
    r.Post("/token/refresh", controllers.RefreshToken)
	r.Post("/question/create", controllers.CreateQuestion)
	r.Get("/question", controllers.GetAllQuestion)
	r.Get("/question/{question_id}", controllers.GetQuestionById)
	r.Delete("/question/{question_id}", controllers.DeleteQuestion)
	r.Patch("/question/{question_id}", controllers.UpdateQuestion)

	return r
}
