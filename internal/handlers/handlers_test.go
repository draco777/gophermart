package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/draco777/gophermart/internal/accrual"
	"github.com/draco777/gophermart/internal/middleware"
	"github.com/draco777/gophermart/internal/models"
	"github.com/draco777/gophermart/internal/service"
	"github.com/draco777/gophermart/internal/storage"
	"github.com/draco777/gophermart/internal/testutil"
)

func setupTestHandlers(t *testing.T) *Handlers {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	balanceStorage := storage.NewBalanceStorage(db)
	accrualClient := &MockAccrualClient{}

	svc := service.NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	return NewHandlers(svc)
}

func TestHandlers_Register(t *testing.T) {
	h := setupTestHandlers(t)

	tests := []struct {
		name           string
		request        models.RegisterRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid registration",
			request: models.RegisterRequest{
				Login:    "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Empty login",
			request: models.RegisterRequest{
				Login:    "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Empty password",
			request: models.RegisterRequest{
				Login:    "testuser",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError {
				// Проверяем, что в ответе есть токен
				authHeader := w.Header().Get("Authorization")
				if authHeader == "" {
					t.Error("Expected Authorization header")
				}

				// Проверяем JSON ответ
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if response["login"] != tt.request.Login {
					t.Errorf("Expected login %s, got %v", tt.request.Login, response["login"])
				}
			}
		})
	}
}

func TestHandlers_Login(t *testing.T) {
	h := setupTestHandlers(t)

	// Сначала регистрируем пользователя
	registerReq := models.RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(registerBody))
	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	h.Register(registerW, registerHTTPReq)

	tests := []struct {
		name           string
		request        models.LoginRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid login",
			request: models.LoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Wrong password",
			request: models.LoginRequest{
				Login:    "testuser",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name: "Non-existing user",
			request: models.LoginRequest{
				Login:    "nonexistent",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError {
				// Проверяем, что в ответе есть токен
				authHeader := w.Header().Get("Authorization")
				if authHeader == "" {
					t.Error("Expected Authorization header")
				}
			}
		})
	}
}

func TestHandlers_AddOrder(t *testing.T) {
	h := setupTestHandlers(t)

	// Создаем пользователя в базе данных
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	userID := testutil.CreateTestUser(t, db, "testuser", "hashedpassword")

	// Создаем контекст с пользователем для тестирования
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.LoginKey, "testuser")

	tests := []struct {
		name           string
		orderNumber    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid order",
			orderNumber:    "12345678903",
			expectedStatus: http.StatusAccepted,
			expectError:    false,
		},
		{
			name:           "Invalid order number",
			orderNumber:    "1234567890",
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(tt.orderNumber)))
			req.Header.Set("Content-Type", "text/plain")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.AddOrder(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandlers_GetOrders(t *testing.T) {
	h := setupTestHandlers(t)

	// Создаем контекст с пользователем для тестирования
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.LoginKey, "testuser")

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.GetOrders(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestHandlers_GetBalance(t *testing.T) {
	h := setupTestHandlers(t)

	// Создаем контекст с пользователем для тестирования
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.LoginKey, "testuser")

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.GetBalance(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var balance models.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &balance); err != nil {
		t.Errorf("Failed to unmarshal balance: %v", err)
	}

	if balance.Current != 0 {
		t.Errorf("Expected current balance 0, got %f", balance.Current)
	}
}

func TestHandlers_WithdrawBalance(t *testing.T) {
	h := setupTestHandlers(t)

	// Создаем контекст с пользователем для тестирования
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.LoginKey, "testuser")

	tests := []struct {
		name           string
		request        models.WithdrawRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Insufficient funds",
			request: models.WithdrawRequest{
				Order: "2377225624",
				Sum:   100.0,
			},
			expectedStatus: http.StatusPaymentRequired,
			expectError:    true,
		},
		{
			name: "Invalid order number",
			request: models.WithdrawRequest{
				Order: "1234567890",
				Sum:   100.0,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.WithdrawBalance(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandlers_GetWithdrawals(t *testing.T) {
	h := setupTestHandlers(t)

	// Создаем контекст с пользователем для тестирования
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.LoginKey, "testuser")

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.GetWithdrawals(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

// MockAccrualClient для тестирования
type MockAccrualClient struct{}

func (m *MockAccrualClient) GetOrderInfo(orderNumber string) (*accrual.AccrualResponse, error) {
	return nil, errors.New("order not found in accrual system")
}
