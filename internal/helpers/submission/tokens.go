package submission

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenManager struct {
	client *redis.Client
}

var Tokens *TokenManager

func Init() error {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	pwd := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		DB:           0,
		Password:     pwd,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     50,
		PoolTimeout:  10 * time.Second,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		fmt.Println("Redis Init Failed: " + err.Error())
		return err
	}
	Tokens = &TokenManager{client: client}
	return nil
}

func (tm *TokenManager) AddToken(ctx context.Context, token string, userID string) error {
	err := tm.client.Set(ctx, fmt.Sprintf("token:%s", token), userID, 0).Err()
	if err != nil {
		return err
	}

	err = tm.client.SAdd(ctx, fmt.Sprintf("user:%s:tokens", userID), token).Err()
	if err != nil {
		return err
	}
	return nil
}

func (tm *TokenManager) GetUserID(ctx context.Context, token string) (string, error) {
	userID, err := tm.client.Get(ctx, fmt.Sprintf("token:%s", token)).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found")
	} else if err != nil {
		return "", err
	}
	return userID, nil
}

func (tm *TokenManager) DeleteToken(ctx context.Context, token string) error {
	userID, err := tm.GetUserID(ctx, token)
	if err != nil {
		return err
	}

	err = tm.client.Del(ctx, fmt.Sprintf("token:%s", token)).Err()
	if err != nil {
		return err
	}

	err = tm.client.SRem(ctx, fmt.Sprintf("user:%s:tokens", userID), token).Err()
	if err != nil {
		return err
	}

	return nil
}
