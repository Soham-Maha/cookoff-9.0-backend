package server

import (
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/controllers"
	helpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"

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

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(helpers.TokenAuth))
		r.Use(jwtauth.Authenticator(helpers.TokenAuth))

		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			httphelpers.WriteJSON(w, http.StatusOK, "Test")
		})
	})

	r.Post("/login/user", controllers.LoginHandler)
	return r
}
