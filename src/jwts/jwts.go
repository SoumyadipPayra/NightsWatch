package jwts

import (
	"errors"
	"time"

	"github.com/SoumyadipPayra/NightsWatch/src/jwts/internal/secrets"
	"github.com/golang-jwt/jwt/v4"
)

func Initialize() error {
	return secrets.Initialize()
}

func GenerateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	secretBytes := secrets.GetSecretKey()
	tokenString, err := token.SignedString(secretBytes)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secrets.GetSecretKey(), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}
	return claims["username"].(string), nil
}
