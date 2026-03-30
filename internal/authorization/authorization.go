package authorization

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/poggerr/gophermart/internal/encrypt"
	"github.com/poggerr/gophermart/internal/logger"
	"github.com/poggerr/gophermart/internal/models"
	"github.com/poggerr/gophermart/internal/storage"
	"os"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID *uuid.UUID
}

const TokenExp = time.Hour * 3

func BuildJWTString(uuid *uuid.UUID) (string, error) {

	var secretKey = os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return "", errors.New("env SECRET_KEY is required")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: uuid,
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GetUserID(tokenString string) *uuid.UUID {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil
	}
	return claims.UserID
}

func RegisterUser(strg *storage.Storage, user *models.User) (uuid.UUID, error) {
	user.Password = encrypt.Encrypt(user.Password)
	id := uuid.New()
	err := strg.CreateUser(user.Username, user.Password, &id)
	if err != nil {
		logger.Initialize().Error(err)
		return id, err
	}
	return id, nil
}

func CheckPass(user *models.User, dbUser *models.User) error {
	decrypted := encrypt.Encrypt(user.Password)
	if dbUser.Password != decrypted {
		return errors.New("ошибка авторизации")
	}
	return nil
}
