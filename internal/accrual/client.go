package accrual

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AccrualClient struct {
	baseURL    string
	httpClient *http.Client
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *AccrualClient) GetOrderInfo(orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &accrualResp, nil

	case http.StatusNoContent:
		return nil, fmt.Errorf("order not found in accrual system")

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		return nil, fmt.Errorf("rate limit exceeded, retry after: %s", retryAfter)

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("internal server error in accrual system")

	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}
