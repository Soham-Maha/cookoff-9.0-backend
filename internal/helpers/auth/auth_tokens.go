package auth

import (
    "context"
    "fmt"
    "os"
    "time"
    "github.com/golang-jwt/jwt/v4"
    logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
    "github.com/redis/go-redis/v9"
)

type AuthTokenManager struct {
    client *redis.Client
	AccessSecret string
}

var Tokens *AuthTokenManager

func Init() {
    host := os.Getenv("DRAGONFLY_HOST")
    port := os.Getenv("DRAGONFLY_PORT")
    pwd := os.Getenv("DRAGONFLY_PASSWORD")

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
        logger.Errof("Redis Init Failed: " + err.Error())
    }
    logger.Infof("Connected to DragonflyDB")
    Tokens = &AuthTokenManager{client: client}
}
func (tm *AuthTokenManager) GenerateAccessToken(userID string) (string, error) {
    claims := jwt.MapClaims{}
    claims["authorized"] = true
    claims["user_id"] = userID
    claims["exp"] = time.Now().Add(time.Minute * 15).Unix() 
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    accessToken, err := token.SignedString([]byte(tm.AccessSecret))
    if err != nil {
        return "", err
    }
	return accessToken,nil
}
func (tm *AuthTokenManager) AddAccessToken(ctx context.Context, accessToken string, userID string) error {
    err := tm.client.Set(ctx, fmt.Sprintf("access_token:%s", accessToken), userID, 24*time.Hour).Err()
    if err != nil {
        return err
    }
    return nil
}

func (tm *AuthTokenManager) GetUserID(ctx context.Context, accessToken string) (string, error) {
    userID, err := tm.client.Get(ctx, fmt.Sprintf("access_token:%s", accessToken)).Result()
    if err == redis.Nil {
        return "", fmt.Errorf("access token not found")
    } else if err != nil {
        return "", err
    }
    return userID, nil
}

func (tm *AuthTokenManager) DeleteAccessToken(ctx context.Context, accessToken string) error {
    err := tm.client.Del(ctx, fmt.Sprintf("access_token:%s", accessToken)).Err()
    if err != nil {
        return err
    }
    return nil
}
