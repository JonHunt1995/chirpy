package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestValidJWT(t *testing.T) {
	// Arrange: Set up our test data
	want := uuid.New()
	tokenSecret := "your-test-secret"
	expiresIn := time.Hour
	// Act - Part 1: Create the token
	token, err := MakeJWT(want, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error creating token: %v", err)
	}
	// Act - Part 2: Try to validate and extract UUID from the token
	got, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Error validating token: %v", err)
	}

	// 4. Assert: Check if the UUID we got back matches that we put in
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "your-test-secret"

	t.Run("negative duration", func(t *testing.T) {
		expiresIn := -time.Hour
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Error creating token: %v", err)
		}
		if _, err = ValidateJWT(token, tokenSecret); err == nil {
			t.Fatalf("Expected validation of expired JWT to fail")
		}
	})

	t.Run("natural expiration", func(t *testing.T) {
		expiresIn := time.Second
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Error creating token: %v", err)
		}
		time.Sleep(2 * time.Second)
		if _, err = ValidateJWT(token, tokenSecret); err == nil {
			t.Fatalf("Expected validation of expired JWT to fail")
		}
	})
}

func TestInvalidSecretJWT(t *testing.T) {
	userId := uuid.New()
	tokenSecret := "correct-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userId, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error creating token: %v", err)
	}
	tokenSecret = "incorrect_secret"
	if _, err = ValidateJWT(token, tokenSecret); err == nil {
		t.Fatalf("Expected validation of JWT with incorrect token secret to fail")
	}
}

func TestInvalidSigningMethod(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "correct_secret"
	expiresIn := time.Hour

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	if _, err = ValidateJWT(tokenString, tokenSecret); err == nil {
		t.Fatalf("Expected validation of JWT with incorrect method to fail")
	}
}
