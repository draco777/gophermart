package accrual

// AccrualClientInterface определяет интерфейс для клиента системы расчета баллов
type AccrualClientInterface interface {
	GetOrderInfo(orderNumber string) (*AccrualResponse, error)
}
