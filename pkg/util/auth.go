package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SECRET = "fuckyou"

func GenerateJWT(raw string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": raw,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString([]byte(SECRET))
}

func ParseJWT(token string) (string, error) {

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	} else {
		return "", errors.New("token format error")
	}

	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(SECRET), nil
	})

	if err != nil {

		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", errors.New("token expired")
		}

		return "", errors.New("invalid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)

	if !ok || !t.Valid {
		return "", errors.New("invalid token")
	}

	userID, ok := claims["user_id"].(string)

	if !ok {
		return "", errors.New("invalid user id")
	}

	return userID, nil
}
