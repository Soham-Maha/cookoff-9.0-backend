package server

import (
	"github.com/CodeChefVIT/cookoff-backend/internal/controllers"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/ping", controllers.HealthCheck)
	r.Post("/submit", controllers.SubmitCode)
	r.Post("/question/create", controllers.CreateQuestion)
	r.Get("/question", controllers.GetAllQuestion)
	r.Get("/question/{question_id}", controllers.GetQuestionById)
	r.Delete("/question/{question_id}", controllers.DeleteQuestion)
	r.Patch("/question/{question_id}", controllers.UpdateQuestion)
	r.Post("/login/user", controllers.LoginHandler)
	r.Post("/token/refresh", controllers.RefreshTokenHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth.TokenAuth))
		r.Use(jwtauth.Authenticator(auth.TokenAuth))

		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			httphelpers.WriteJSON(w, http.StatusOK, "Test")
		})
	})

	return r
}
