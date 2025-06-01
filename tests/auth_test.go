package tests

import (
	"testing"
	"time"

	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app/handlers"
)

func TestHashPasswordPositive(t *testing.T) {
	password := "12345"

	hashedPassword, err := handlers.HashPassword(password)
	if err != nil {
		t.Fatalf("error hashing password")
	}

	if compareErr := handlers.CheckPasswordHash(hashedPassword, password); compareErr != nil {
		t.Fatalf("passwords do not match")
	}
}

func TestHashPasswordNegative(t *testing.T) {
	password := "12345"
	wrongPassword := "abcd"

	hashedPassword, err := handlers.HashPassword(password)
	if err != nil {
		t.Fatalf("error hashing password")
	}

	if compareErr := handlers.CheckPasswordHash(hashedPassword, wrongPassword); compareErr == nil {
		t.Fatalf("no errors are thrown for passwords not matching.")
	}
}

func TestMakeJWT(t *testing.T) {
	userId := int64(22)
	tokenSecret := "rT8!pL3#vQ7@kF9$mZ2&xY5*wC1^"

	_, err := handlers.MakeJWT(userId, tokenSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("error making jwt token")
	}
}

func TestValidateJWT(t *testing.T) {
	userId := int64(24)
	tokenSecret := "rT8!pL3#vQ7@kF9$mZ2&xY5*wC1^"

	signedJWTString, err := handlers.MakeJWT(userId, tokenSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("error making jwt token")
	}

	validatedUserId, validateErr := handlers.ValidateJWT(signedJWTString, tokenSecret)
	if validateErr != nil {
		t.Fatalf("error validating jwt : %v", validateErr)
	}

	if userId != validatedUserId {
		t.Fatalf("user ids do not match")
	}
}
