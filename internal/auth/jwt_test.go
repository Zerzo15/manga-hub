package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	service, err := NewJWTService()
	if err != nil {
		t.Fatal(err)
	}

	token, err := service.GenerateToken(42, "alice")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || claims.Username != "alice" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "first-secret")
	service, err := NewJWTService()
	if err != nil {
		t.Fatal(err)
	}

	token, err := service.GenerateToken(42, "alice")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("JWT_SECRET", "second-secret")
	otherService, err := NewJWTService()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := otherService.ValidateToken(token); err == nil {
		t.Fatal("expected wrong-secret token to fail validation")
	}
}

func TestValidateTokenRejectsExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	service, err := NewJWTService()
	if err != nil {
		t.Fatal(err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:   42,
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	tokenString, err := token.SignedString(service.secret)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.ValidateToken(tokenString); err == nil {
		t.Fatal("expected expired token to fail validation")
	}
}

func TestValidateTokenRejectsInvalidSigningAlgorithm(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	service, err := NewJWTService()
	if err != nil {
		t.Fatal(err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, Claims{
		UserID:   42,
		Username: "alice",
	})
	tokenString, err := token.SignedString(service.secret)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.ValidateToken(tokenString); err == nil {
		t.Fatal("expected invalid signing algorithm to fail validation")
	}
}

func TestValidateTokenRejectsNonPositiveUserID(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	service, err := NewJWTService()
	if err != nil {
		t.Fatal(err)
	}

	token, err := service.GenerateToken(0, "alice")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.ValidateToken(token); err == nil {
		t.Fatal("expected non-positive user ID to fail validation")
	}
}
