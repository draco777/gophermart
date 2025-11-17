package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)

	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not be the same as password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "Correct password",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "Wrong password",
			password: "wrongpassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPasswordHash(tt.password, tt.hash)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	userID := 123
	login := "testuser"

	token, err := GenerateToken(userID, login)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Проверяем, что токен можно валидировать
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Login != login {
		t.Errorf("Expected Login %s, got %s", login, claims.Login)
	}
}

func TestValidateToken(t *testing.T) {
	userID := 456
	login := "testuser2"

	// Генерируем валидный токен
	token, err := GenerateToken(userID, login)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		expectErr bool
	}{
		{
			name:      "Valid token",
			token:     token,
			expectErr: false,
		},
		{
			name:      "Invalid token",
			token:     "invalid.token.here",
			expectErr: true,
		},
		{
			name:      "Empty token",
			token:     "",
			expectErr: true,
		},
		{
			name:      "Malformed token",
			token:     "not.a.token",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if claims.UserID != userID {
					t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
				}

				if claims.Login != login {
					t.Errorf("Expected Login %s, got %s", login, claims.Login)
				}
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	userID := 789
	login := "testuser3"

	// Создаем токен с коротким временем жизни для тестирования
	claims := Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Истек час назад
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Пытаемся валидировать истекший токен
	_, err = ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"Length 10", 10},
		{"Length 20", 20},
		{"Length 32", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := GenerateRandomString(tt.length)
			if err != nil {
				t.Fatalf("GenerateRandomString failed: %v", err)
			}

			// Проверяем, что длина строки в два раза больше (hex encoding)
			expectedLength := tt.length * 2
			if len(str) != expectedLength {
				t.Errorf("Expected length %d, got %d", expectedLength, len(str))
			}

			// Проверяем, что строка не пустая
			if str == "" {
				t.Error("Generated string should not be empty")
			}
		})
	}
}
