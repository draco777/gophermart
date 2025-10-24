package storage

import (
	"database/sql"
	"fmt"

	"github.com/draco777/gophermart/internal/models"
)

type OrderStorage struct {
	db *sql.DB
}

func NewOrderStorage(db *sql.DB) *OrderStorage {
	return &OrderStorage{db: db}
}

func (s *OrderStorage) CreateOrder(userID int, number string) (*models.Order, error) {
	query := `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3) RETURNING id, uploaded_at`

	var order models.Order
	err := s.db.QueryRow(query, userID, number, models.OrderStatusNew).Scan(&order.ID, &order.UploadedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	order.UserID = userID
	order.Number = number
	order.Status = models.OrderStatusNew

	return &order, nil
}

func (s *OrderStorage) GetOrdersByUserID(userID int) ([]models.OrderResponse, error) {
	query := `SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []models.OrderResponse
	for rows.Next() {
		var order models.OrderResponse
		var accrual sql.NullFloat64

		err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if accrual.Valid {
			order.Accrual = accrual.Float64
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (s *OrderStorage) GetOrderByNumber(number string) (*models.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE number = $1`

	var order models.Order
	var accrual sql.NullFloat64

	err := s.db.QueryRow(query, number).Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &accrual, &order.UploadedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if accrual.Valid {
		order.Accrual = accrual.Float64
	}

	return &order, nil
}

func (s *OrderStorage) UpdateOrderStatus(number, status string, accrual float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`

	result, err := s.db.Exec(query, status, accrual, number)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func (s *OrderStorage) GetProcessingOrders() ([]models.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE status IN ($1, $2)`

	rows, err := s.db.Query(query, models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var accrual sql.NullFloat64

		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if accrual.Valid {
			order.Accrual = accrual.Float64
		}

		orders = append(orders, order)
	}

	return orders, nil
}
