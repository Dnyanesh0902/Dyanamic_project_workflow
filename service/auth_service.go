package service

import (
	"errors"
	"fmt"
	"os"
	"project-workflow-backend/model"
	"time"

	"github.com/golang-jwt/jwt"
)

// GenerateToken generates a JWT token for a given user using the Secret Key
func GenerateToken(user *model.User) (string, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "default_secret_key"
	}

	claims := jwt.MapClaims{
		"id":        user.ID,
		"uuid":      user.UUID,
		"user_type": user.Rolename,
		"pincode":   "", // Expected by token middleware
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the parsed token
func ValidateToken(tokenString string) (*jwt.Token, error) {
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "default_secret_key"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
