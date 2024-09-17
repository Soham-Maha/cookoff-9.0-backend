package server

import (
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/controllers"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/hibiken/asynq"
)

func (s *Server) RegisterRoutes(taskClient *asynq.Client) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/ping", controllers.HealthCheck)
	r.Put("/callback", func(w http.ResponseWriter, r *http.Request) {
		controllers.CallbackUrl(w, r, taskClient)
	})
	r.Post("/question/create", controllers.CreateQuestion)
	r.Get("/question", controllers.GetAllQuestion)
	r.Get("/question/{question_id}", controllers.GetQuestionById)
	r.Delete("/question/{question_id}", controllers.DeleteQuestion)
	r.Patch("/question/{question_id}", controllers.UpdateQuestion)
	r.Post("/login/user", controllers.LoginHandler)
	r.Post("/token/refresh", controllers.RefreshTokenHandler)
	r.Group(func(protected chi.Router) {
		protected.Use(jwtauth.Verifier(auth.TokenAuth))
		protected.Use(jwtauth.Authenticator(auth.TokenAuth))

		protected.Get("/protected", controllers.ProtectedHandler)
		protected.Post("/submit", controllers.SubmitCode)
		protected.Post("/runcode", controllers.RunCode)
	})

	return r
}
