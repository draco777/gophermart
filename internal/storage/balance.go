package storage

import (
	"database/sql"
	"fmt"

	"github.com/draco777/gophermart/internal/models"
)

type BalanceStorage struct {
	db *sql.DB
}

func NewBalanceStorage(db *sql.DB) *BalanceStorage {
	return &BalanceStorage{db: db}
}

func (s *BalanceStorage) GetBalance(userID int) (*models.BalanceResponse, error) {
	query := `SELECT current, withdrawn FROM balances WHERE user_id = $1`

	var balance models.BalanceResponse
	err := s.db.QueryRow(query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("balance not found")
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &balance, nil
}

func (s *BalanceStorage) UpdateBalance(userID int, accrual float64) error {
	query := `UPDATE balances SET current = current + $1, updated_at = NOW() WHERE user_id = $2`

	result, err := s.db.Exec(query, accrual, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("balance not found")
	}

	return nil
}

func (s *BalanceStorage) WithdrawBalance(userID int, amount float64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем достаточность средств
	var currentBalance float64
	err = tx.QueryRow(`SELECT current FROM balances WHERE user_id = $1`, userID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient funds")
	}

	// Обновляем баланс
	_, err = tx.Exec(`UPDATE balances SET current = current - $1, withdrawn = withdrawn + $1, updated_at = NOW() WHERE user_id = $2`,
		amount, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return tx.Commit()
}

func (s *BalanceStorage) CreateWithdrawal(userID int, orderNumber string, amount float64) error {
	query := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`

	_, err := s.db.Exec(query, userID, orderNumber, amount)
	if err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return nil
}

func (s *BalanceStorage) GetWithdrawals(userID int) ([]models.WithdrawalResponse, error) {
	query := `SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []models.WithdrawalResponse
	for rows.Next() {
		var withdrawal models.WithdrawalResponse

		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate withdrawals: %w", err)
	}

	return withdrawals, nil
}
