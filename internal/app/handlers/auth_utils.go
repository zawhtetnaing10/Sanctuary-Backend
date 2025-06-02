package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password %w", err)
	}

	return string(hashedBytes), nil
}

func CheckPasswordHash(hash, password string) error {
	if len(hash) == 0 {
		return fmt.Errorf("password hash cannot be empty")
	}

	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// JWT
func MakeJWT(id int64, tokenSecret string, expiresIn time.Duration) (string, error) {
	if id == 0 {
		return "", fmt.Errorf("the id must not be 0")
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Issuer:    "sanctuary",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   strconv.FormatInt(id, 10),
		})

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("error signing token %w", err)
	}

	return signedToken, nil
}

// Validate JWT Token
func ValidateJWT(tokenString, tokenSecret string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(tokenSecret), nil
	})
	if err != nil {
		return 0, err
	}

	userIdString, userIdErr := token.Claims.GetSubject()
	if userIdErr != nil {
		return 0, userIdErr
	}

	userId, parseErr := strconv.ParseInt(userIdString, 10, 64)
	if parseErr != nil {
		return 0, fmt.Errorf("error parsing user id : %w", parseErr)
	}
	return userId, nil
}

// Get Bearer Token
func GetBearerToken(headers http.Header) (string, error) {
	// Get bearer token
	authHeader := headers.Get(app.AUTHORIZATION)
	// Auth token must not be empty
	if authHeader == "" {
		return "", errors.New("the auth token must not be empty")
	}

	// Remove prefix Bearer
	tokenString := strings.TrimPrefix(authHeader, app.BEAERER)

	// user sends in only Bearer with no token string
	if tokenString == authHeader {
		return "", errors.New("invalid bearer token format. The correct format is Bearer {token}")
	}

	// Check for empty token
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", errors.New("the token string must not be empty")
	}

	return tokenString, nil
}
