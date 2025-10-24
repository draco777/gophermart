package models

import (
	"testing"
	"time"
)

func TestOrderStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"Valid NEW status", OrderStatusNew, true},
		{"Valid PROCESSING status", OrderStatusProcessing, true},
		{"Valid INVALID status", OrderStatusInvalid, true},
		{"Valid PROCESSED status", OrderStatusProcessed, true},
		{"Invalid status", "UNKNOWN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.status == OrderStatusNew ||
				tt.status == OrderStatusProcessing ||
				tt.status == OrderStatusInvalid ||
				tt.status == OrderStatusProcessed

			if valid != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, valid)
			}
		})
	}
}

func TestRegisterRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  RegisterRequest
		expected bool
	}{
		{
			name: "Valid request",
			request: RegisterRequest{
				Login:    "testuser",
				Password: "password123",
			},
			expected: true,
		},
		{
			name: "Empty login",
			request: RegisterRequest{
				Login:    "",
				Password: "password123",
			},
			expected: false,
		},
		{
			name: "Empty password",
			request: RegisterRequest{
				Login:    "testuser",
				Password: "",
			},
			expected: false,
		},
		{
			name: "Both empty",
			request: RegisterRequest{
				Login:    "",
				Password: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.request.Login != "" && tt.request.Password != ""
			if valid != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, valid)
			}
		})
	}
}

func TestWithdrawRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  WithdrawRequest
		expected bool
	}{
		{
			name: "Valid request",
			request: WithdrawRequest{
				Order: "12345678903",
				Sum:   100.50,
			},
			expected: true,
		},
		{
			name: "Empty order",
			request: WithdrawRequest{
				Order: "",
				Sum:   100.50,
			},
			expected: false,
		},
		{
			name: "Zero sum",
			request: WithdrawRequest{
				Order: "12345678903",
				Sum:   0,
			},
			expected: false,
		},
		{
			name: "Negative sum",
			request: WithdrawRequest{
				Order: "12345678903",
				Sum:   -10,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.request.Order != "" && tt.request.Sum > 0
			if valid != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, valid)
			}
		})
	}
}

func TestOrderResponse(t *testing.T) {
	now := time.Now()
	order := OrderResponse{
		Number:     "12345678903",
		Status:     OrderStatusProcessed,
		Accrual:    500.0,
		UploadedAt: now,
	}

	if order.Number != "12345678903" {
		t.Errorf("Expected number 12345678903, got %s", order.Number)
	}

	if order.Status != OrderStatusProcessed {
		t.Errorf("Expected status PROCESSED, got %s", order.Status)
	}

	if order.Accrual != 500.0 {
		t.Errorf("Expected accrual 500.0, got %f", order.Accrual)
	}

	if !order.UploadedAt.Equal(now) {
		t.Errorf("Expected uploaded_at to be %v, got %v", now, order.UploadedAt)
	}
}

func TestBalanceResponse(t *testing.T) {
	balance := BalanceResponse{
		Current:   1000.50,
		Withdrawn: 250.75,
	}

	if balance.Current != 1000.50 {
		t.Errorf("Expected current 1000.50, got %f", balance.Current)
	}

	if balance.Withdrawn != 250.75 {
		t.Errorf("Expected withdrawn 250.75, got %f", balance.Withdrawn)
	}
}

func TestWithdrawalResponse(t *testing.T) {
	now := time.Now()
	withdrawal := WithdrawalResponse{
		Order:       "2377225624",
		Sum:         500.0,
		ProcessedAt: now,
	}

	if withdrawal.Order != "2377225624" {
		t.Errorf("Expected order 2377225624, got %s", withdrawal.Order)
	}

	if withdrawal.Sum != 500.0 {
		t.Errorf("Expected sum 500.0, got %f", withdrawal.Sum)
	}

	if !withdrawal.ProcessedAt.Equal(now) {
		t.Errorf("Expected processed_at to be %v, got %v", now, withdrawal.ProcessedAt)
	}
}
