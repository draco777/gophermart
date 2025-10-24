package service

import (
	"errors"
	"testing"

	"github.com/draco777/gophermart/internal/accrual"
	"github.com/draco777/gophermart/internal/storage"
	"github.com/draco777/gophermart/internal/testutil"
)

// MockAccrualClient для тестирования
type MockAccrualClient struct {
	responses map[string]*accrual.AccrualResponse
	errors    map[string]error
}

func NewMockAccrualClient() *MockAccrualClient {
	return &MockAccrualClient{
		responses: make(map[string]*accrual.AccrualResponse),
		errors:    make(map[string]error),
	}
}

func (m *MockAccrualClient) SetResponse(orderNumber string, response *accrual.AccrualResponse) {
	m.responses[orderNumber] = response
}

func (m *MockAccrualClient) SetError(orderNumber string, err error) {
	m.errors[orderNumber] = err
}

func (m *MockAccrualClient) GetOrderInfo(orderNumber string) (*accrual.AccrualResponse, error) {
	if err, exists := m.errors[orderNumber]; exists {
		return nil, err
	}

	if response, exists := m.responses[orderNumber]; exists {
		return response, nil
	}

	return nil, errors.New("order not found in accrual system")
}

func TestService_RegisterUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	balanceStorage := storage.NewBalanceStorage(db)
	accrualClient := NewMockAccrualClient()

	service := NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	tests := []struct {
		name        string
		login       string
		password    string
		expectError bool
	}{
		{
			name:        "Valid registration",
			login:       "testuser",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "Empty login",
			login:       "",
			password:    "password123",
			expectError: false, // Сервис не проверяет пустой логин на уровне валидации
		},
		{
			name:        "Empty password",
			login:       "testuser",
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := service.RegisterUser(tt.login, tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if user.Login != tt.login {
					t.Errorf("Expected login %s, got %s", tt.login, user.Login)
				}

				if token == "" {
					t.Error("Token should not be empty")
				}
			}
		})
	}
}

func TestService_LoginUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	balanceStorage := storage.NewBalanceStorage(db)
	accrualClient := NewMockAccrualClient()

	service := NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	// Создаем тестового пользователя
	_, _, err := service.RegisterUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name        string
		login       string
		password    string
		expectError bool
	}{
		{
			name:        "Valid login",
			login:       "testuser",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "Wrong password",
			login:       "testuser",
			password:    "wrongpassword",
			expectError: true,
		},
		{
			name:        "Non-existing user",
			login:       "nonexistent",
			password:    "password123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := service.LoginUser(tt.login, tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if user.Login != tt.login {
					t.Errorf("Expected login %s, got %s", tt.login, user.Login)
				}

				if token == "" {
					t.Error("Token should not be empty")
				}
			}
		})
	}
}

func TestService_AddOrder(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	balanceStorage := storage.NewBalanceStorage(db)
	accrualClient := NewMockAccrualClient()

	service := NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	// Создаем тестового пользователя
	user, _, err := service.RegisterUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name        string
		userID      int
		orderNumber string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid order",
			userID:      user.ID,
			orderNumber: "12345678903",
			expectError: false,
		},
		{
			name:        "Invalid order number",
			userID:      user.ID,
			orderNumber: "1234567890",
			expectError: true,
			errorMsg:    "invalid order number format",
		},
		{
			name:        "Empty order number",
			userID:      user.ID,
			orderNumber: "",
			expectError: true,
			errorMsg:    "invalid order number format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AddOrder(tt.userID, tt.orderNumber)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_GetUserOrders(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	balanceStorage := storage.NewBalanceStorage(db)
	accrualClient := NewMockAccrualClient()

	service := NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	// Создаем тестового пользователя
	user, _, err := service.RegisterUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Добавляем тестовые заказы
	err = service.AddOrder(user.ID, "12345678903")
	if err != nil {
		t.Fatalf("Failed to add test order: %v", err)
	}

	err = service.AddOrder(user.ID, "9278923470")
	if err != nil {
		t.Fatalf("Failed to add test order: %v", err)
	}

	orders, err := service.GetUserOrders(user.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}

	// Проверяем, что заказы отсортированы по времени
	if orders[0].Number != "9278923470" {
		t.Errorf("Expected first order to be 9278923470, got %s", orders[0].Number)
	}

	if orders[1].Number != "12345678903" {
		t.Errorf("Expected second order to be 12345678903, got %s", orders[1].Number)
	}
}

func TestService_GetUserBalance(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	balanceStorage := storage.NewBalanceStorage(db)
	accrualClient := NewMockAccrualClient()

	service := NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	// Создаем тестового пользователя
	user, _, err := service.RegisterUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	balance, err := service.GetUserBalance(user.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if balance.Current != 0 {
		t.Errorf("Expected current balance 0, got %f", balance.Current)
	}

	if balance.Withdrawn != 0 {
		t.Errorf("Expected withdrawn balance 0, got %f", balance.Withdrawn)
	}
}

func TestService_WithdrawBalance(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	balanceStorage := storage.NewBalanceStorage(db)
	accrualClient := NewMockAccrualClient()

	service := NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	// Создаем тестового пользователя
	user, _, err := service.RegisterUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Устанавливаем баланс
	err = balanceStorage.UpdateBalance(user.ID, 1000.0)
	if err != nil {
		t.Fatalf("Failed to update balance: %v", err)
	}

	tests := []struct {
		name        string
		userID      int
		orderNumber string
		amount      float64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid withdrawal",
			userID:      user.ID,
			orderNumber: "2377225624",
			amount:      500.0,
			expectError: false,
		},
		{
			name:        "Insufficient funds",
			userID:      user.ID,
			orderNumber: "2377225624",
			amount:      1500.0,
			expectError: true,
			errorMsg:    "insufficient funds",
		},
		{
			name:        "Invalid order number",
			userID:      user.ID,
			orderNumber: "1234567890",
			amount:      100.0,
			expectError: true,
			errorMsg:    "invalid order number format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.WithdrawBalance(tt.userID, tt.orderNumber, tt.amount)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_ValidateOrderNumber(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	balanceStorage := storage.NewBalanceStorage(db)
	accrualClient := NewMockAccrualClient()

	service := NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	tests := []struct {
		name        string
		orderNumber string
		expected    bool
	}{
		{
			name:        "Valid order number",
			orderNumber: "12345678903",
			expected:    true,
		},
		{
			name:        "Valid order number 2",
			orderNumber: "9278923470",
			expected:    true,
		},
		{
			name:        "Invalid order number",
			orderNumber: "1234567890",
			expected:    false,
		},
		{
			name:        "Empty order number",
			orderNumber: "",
			expected:    false,
		},
		{
			name:        "Non-numeric order number",
			orderNumber: "abc123",
			expected:    false,
		},
		{
			name:        "Too short order number",
			orderNumber: "123",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateOrderNumber(tt.orderNumber)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v for order number %s", tt.expected, result, tt.orderNumber)
			}
		})
	}
}
