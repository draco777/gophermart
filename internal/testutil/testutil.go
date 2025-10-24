package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const TestDatabaseURL = "postgres://svyaz_user:svyaz_password@localhost:5432/gophermart_test?sslmode=disable"

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Проверяем переменную окружения для тестовой БД
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = TestDatabaseURL
	}

	db, err := sql.Open("postgres", testDBURL)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Запускаем миграции для тестовой БД
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("Failed to set dialect: %v", err)
	}

	if err := goose.Up(db, "../../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	// Очищаем все таблицы
	tables := []string{"withdrawals", "balances", "orders", "users"}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Failed to clean table %s: %v", table, err)
		}
	}

	db.Close()
}

func CreateTestUser(t *testing.T, db *sql.DB, login, password string) int {
	t.Helper()

	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`

	var userID int
	err := db.QueryRow(query, login, password).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Создаем баланс для пользователя
	_, err = db.Exec(`INSERT INTO balances (user_id) VALUES ($1)`, userID)
	if err != nil {
		t.Fatalf("Failed to create balance for test user: %v", err)
	}

	return userID
}

func CreateTestOrder(t *testing.T, db *sql.DB, userID int, number, status string, accrual float64) int {
	t.Helper()

	query := `INSERT INTO orders (user_id, number, status, accrual) VALUES ($1, $2, $3, $4) RETURNING id`

	var orderID int
	err := db.QueryRow(query, userID, number, status, accrual).Scan(&orderID)
	if err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}

	return orderID
}

func CreateTestWithdrawal(t *testing.T, db *sql.DB, userID int, orderNumber string, sum float64) int {
	t.Helper()

	query := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3) RETURNING id`

	var withdrawalID int
	err := db.QueryRow(query, userID, orderNumber, sum).Scan(&withdrawalID)
	if err != nil {
		t.Fatalf("Failed to create test withdrawal: %v", err)
	}

	return withdrawalID
}

func UpdateTestBalance(t *testing.T, db *sql.DB, userID int, current, withdrawn float64) {
	t.Helper()

	query := `UPDATE balances SET current = $1, withdrawn = $2 WHERE user_id = $3`

	_, err := db.Exec(query, current, withdrawn, userID)
	if err != nil {
		t.Fatalf("Failed to update test balance: %v", err)
	}
}
