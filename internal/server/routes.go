package server

import (
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/controllers"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/CodeChefVIT/cookoff-backend/internal/middlewares"
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
	r.Get("/testcase", controllers.GetAllTestCasesHandler)

	r.Post("/user/signup", controllers.SignUp)

	r.Post("/login/user", controllers.LoginHandler)
	r.Post("/token/refresh", controllers.RefreshTokenHandler)
	r.Group(func(protected chi.Router) {
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Post("/upgrade", controllers.UpgradeUserToRound)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Post("/roast", controllers.BanUser)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Post("/unroast", controllers.UnbanUser)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Post("/round/", controllers.SetRoundStatus)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Get("/users", controllers.GetAllUsers)
		protected.Use(jwtauth.Verifier(auth.TokenAuth))
		protected.Use(jwtauth.Authenticator(auth.TokenAuth))

		protected.Get("/me", controllers.MeHandler)
		protected.Get("/protected", controllers.ProtectedHandler)
		protected.With(middlewares.BanCheckMiddleware).Post("/submit", controllers.SubmitCode)
        protected.With(middlewares.BanCheckMiddleware).Post("/runcode", controllers.RunCode)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Post("/question/create", controllers.CreateQuestion)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Get("/questions", controllers.GetAllQuestion)
		protected.With(middlewares.BanCheckMiddleware).Get("/question/round", controllers.GetQuestionsByRound)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Get("/question/{question_id}", controllers.GetQuestionById)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Delete("/question/{question_id}", controllers.DeleteQuestion)
		protected.With(middlewares.RoleAuthorizationMiddleware("admin")).Patch("/question", controllers.UpdateQuestion)
	})

	r.Group(func(protected chi.Router) {
		protected.Use(jwtauth.Verifier(auth.TokenAuth))
		protected.Use(jwtauth.Authenticator(auth.TokenAuth))
		protected.Use(middlewares.RoleAuthorizationMiddleware("admin"))

		protected.Post("/testcase", controllers.CreateTestCaseHandler)
		protected.Put("/testcase/{testcase_id}", controllers.UpdateTestCaseHandler)
		protected.Delete("/testcase/{testcase_id}", controllers.DeleteTestCaseHandler)
		protected.Get("/testcase/{testcase_id}", controllers.GetTestCaseHandler)
	})

	return r
}
