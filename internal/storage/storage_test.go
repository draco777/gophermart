package storage

import (
	"testing"

	"github.com/draco777/gophermart/internal/models"
	"github.com/draco777/gophermart/internal/testutil"
)

func TestUserStorage_CreateUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := NewUserStorage(db)

	tests := []struct {
		name         string
		login        string
		passwordHash string
		expectError  bool
	}{
		{
			name:         "Valid user",
			login:        "testuser",
			passwordHash: "hashedpassword",
			expectError:  false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Добавляем уникальный суффикс для избежания дублирования
			login := tt.login
			if login != "" {
				login = tt.login + "_" + string(rune('0'+i))
			}

			user, err := userStorage.CreateUser(login, tt.passwordHash)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if user == nil {
					t.Error("User should not be nil")
					return
				}

				if user.Login != login {
					t.Errorf("Expected login %s, got %s", login, user.Login)
				}

				if user.PasswordHash != tt.passwordHash {
					t.Errorf("Expected password hash %s, got %s", tt.passwordHash, user.PasswordHash)
				}

				if user.ID == 0 {
					t.Error("User ID should not be zero")
				}
			}
		})
	}
}

func TestUserStorage_GetUserByLogin(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := NewUserStorage(db)

	// Создаем тестового пользователя
	userID := testutil.CreateTestUser(t, db, "testuser", "hashedpassword")

	tests := []struct {
		name        string
		login       string
		expectError bool
	}{
		{
			name:        "Existing user",
			login:       "testuser",
			expectError: false,
		},
		{
			name:        "Non-existing user",
			login:       "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userStorage.GetUserByLogin(tt.login)

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

				if user.ID != userID {
					t.Errorf("Expected user ID %d, got %d", userID, user.ID)
				}
			}
		})
	}
}

func TestUserStorage_UserExists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	userStorage := NewUserStorage(db)

	// Создаем тестового пользователя
	testutil.CreateTestUser(t, db, "testuser", "hashedpassword")

	tests := []struct {
		name     string
		login    string
		expected bool
	}{
		{
			name:     "Existing user",
			login:    "testuser",
			expected: true,
		},
		{
			name:     "Non-existing user",
			login:    "nonexistent",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := userStorage.UserExists(tt.login)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if exists != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, exists)
			}
		})
	}
}

func TestOrderStorage_CreateOrder(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	orderStorage := NewOrderStorage(db)
	userID := testutil.CreateTestUser(t, db, "testuser", "hashedpassword")

	tests := []struct {
		name        string
		userID      int
		number      string
		expectError bool
	}{
		{
			name:        "Valid order",
			userID:      userID,
			number:      "12345678903",
			expectError: false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Добавляем уникальный суффикс для избежания дублирования
			number := tt.number
			if number != "" {
				number = tt.number + "_" + string(rune('0'+i))
			}

			order, err := orderStorage.CreateOrder(tt.userID, number)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if order == nil {
					t.Error("Order should not be nil")
					return
				}

				if order.UserID != tt.userID {
					t.Errorf("Expected user ID %d, got %d", tt.userID, order.UserID)
				}

				if order.Number != number {
					t.Errorf("Expected number %s, got %s", number, order.Number)
				}

				if order.Status != models.OrderStatusNew {
					t.Errorf("Expected status %s, got %s", models.OrderStatusNew, order.Status)
				}
			}
		})
	}
}

func TestOrderStorage_GetOrdersByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	orderStorage := NewOrderStorage(db)
	userID := testutil.CreateTestUser(t, db, "testuser", "hashedpassword")

	// Создаем тестовые заказы
	testutil.CreateTestOrder(t, db, userID, "12345678903", models.OrderStatusNew, 0)
	testutil.CreateTestOrder(t, db, userID, "9278923470", models.OrderStatusProcessed, 500)

	orders, err := orderStorage.GetOrdersByUserID(userID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}

	// Проверяем, что заказы отсортированы по времени (новые первые)
	if orders[0].Number != "9278923470" {
		t.Errorf("Expected first order to be 9278923470, got %s", orders[0].Number)
	}

	if orders[1].Number != "12345678903" {
		t.Errorf("Expected second order to be 12345678903, got %s", orders[1].Number)
	}
}

func TestBalanceStorage_GetBalance(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	balanceStorage := NewBalanceStorage(db)
	userID := testutil.CreateTestUser(t, db, "testuser", "hashedpassword")

	// Устанавливаем тестовый баланс
	testutil.UpdateTestBalance(t, db, userID, 1000.50, 250.75)

	balance, err := balanceStorage.GetBalance(userID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if balance.Current != 1000.50 {
		t.Errorf("Expected current 1000.50, got %f", balance.Current)
	}

	if balance.Withdrawn != 250.75 {
		t.Errorf("Expected withdrawn 250.75, got %f", balance.Withdrawn)
	}
}

func TestBalanceStorage_UpdateBalance(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	balanceStorage := NewBalanceStorage(db)
	userID := testutil.CreateTestUser(t, db, "testuser", "hashedpassword")

	// Обновляем баланс
	err := balanceStorage.UpdateBalance(userID, 500.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем результат
	balance, err := balanceStorage.GetBalance(userID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if balance.Current != 500.0 {
		t.Errorf("Expected current 500.0, got %f", balance.Current)
	}
}

func TestBalanceStorage_WithdrawBalance(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	balanceStorage := NewBalanceStorage(db)
	userID := testutil.CreateTestUser(t, db, "testuser", "hashedpassword")

	// Устанавливаем начальный баланс
	testutil.UpdateTestBalance(t, db, userID, 1000.0, 0)

	tests := []struct {
		name        string
		amount      float64
		expectError bool
	}{
		{
			name:        "Valid withdrawal",
			amount:      500.0,
			expectError: false,
		},
		{
			name:        "Insufficient funds",
			amount:      1500.0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := balanceStorage.WithdrawBalance(userID, tt.amount)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
