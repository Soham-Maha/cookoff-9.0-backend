package main

import (
	"fmt"

	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	database "github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/submission"
	"github.com/CodeChefVIT/cookoff-backend/internal/server"
)

func main() {
	logger.InitLogger()
	database.Init()
	database.InitCache()
	auth.InitJWT()
	submission.Init(database.RedisClient)
	server := server.NewServer()

	logger.Infof("Server started at port 8080")

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
