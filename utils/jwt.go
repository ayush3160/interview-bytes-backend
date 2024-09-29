package utils

import (
	"os"

	"github.com/ayush3160/interview-bytes-backend/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtSecretKey string

func init() {
	jwtSecretKey = os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		jwtSecretKey = "secret-dafault"
	}
}

func CreateJWT(username string, Id primitive.ObjectID) (string, error) {
	claims := models.CustomClaims{
		Username: username,
		ID:       Id.Hex(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtSecretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseJWT(tokenString string) (*models.CustomClaims, error) {
	claims := &models.CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
