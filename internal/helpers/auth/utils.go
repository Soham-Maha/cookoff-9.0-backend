package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func GetUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		httphelpers.WriteError(w, http.StatusUnauthorized, fmt.Sprintf("unauthorized: %v", err))
		return uuid.UUID{}, err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		httphelpers.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized: user_id not found in claims",
		)
		return uuid.UUID{}, fmt.Errorf("user_id not found in claims")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		httphelpers.WriteError(w, http.StatusUnauthorized, fmt.Sprintf("unauthorized: %v", err))
		return uuid.UUID{}, err
	}

	return userUUID, nil
}

func VerifyRound(ctx context.Context, userID uuid.UUID, questionID uuid.UUID) (bool, error) {
	user, err := database.Queries.GetUserById(ctx, userID)
	if err != nil {
		return false, err
	}

	question, err := database.Queries.GetQuestion(ctx, questionID)
	if err != nil {
		return false, err
	}

	return user.RoundQualified == question.Round, nil
}

func RefreshTokenExist(ctx context.Context, userid string) (bool, error) {
	_, err := database.RedisClient.Get(ctx, userid).Result()

	if errors.Is(err, redis.Nil) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	return false, nil
}

func CheckRefreshToken(ctx context.Context, userid string, token string) (bool, error) {
	cacheToken, err := database.RedisClient.Get(ctx, userid).Result()
	if err != nil {
		return false, fmt.Errorf("error while matching token from cache %v", err.Error())
	}

	return token == cacheToken, nil
}
