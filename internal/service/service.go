package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/draco777/gophermart/internal/accrual"
	"github.com/draco777/gophermart/internal/auth"
	"github.com/draco777/gophermart/internal/models"
	"github.com/draco777/gophermart/internal/storage"
)

type Service struct {
	userStorage    *storage.UserStorage
	orderStorage   *storage.OrderStorage
	balanceStorage *storage.BalanceStorage
	accrualClient  accrual.AccrualClientInterface
}

func NewService(userStorage *storage.UserStorage, orderStorage *storage.OrderStorage,
	balanceStorage *storage.BalanceStorage, accrualClient accrual.AccrualClientInterface) *Service {
	return &Service{
		userStorage:    userStorage,
		orderStorage:   orderStorage,
		balanceStorage: balanceStorage,
		accrualClient:  accrualClient,
	}
}

func (s *Service) RegisterUser(login, password string) (*models.User, string, error) {
	// Проверяем, что пользователь не существует
	exists, err := s.userStorage.UserExists(login)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, "", fmt.Errorf("user already exists")
	}

	// Хешируем пароль
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user, err := s.userStorage.CreateUser(login, passwordHash)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токен
	token, err := auth.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *Service) LoginUser(login, password string) (*models.User, string, error) {
	// Получаем пользователя
	user, err := s.userStorage.GetUserByLogin(login)
	if err != nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Проверяем пароль
	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Генерируем токен
	token, err := auth.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *Service) AddOrder(userID int, orderNumber string) error {
	// Проверяем формат номера заказа (алгоритм Луна)
	if !s.validateOrderNumber(orderNumber) {
		return fmt.Errorf("invalid order number format")
	}

	// Проверяем, не существует ли уже такой заказ
	existingOrder, err := s.orderStorage.GetOrderByNumber(orderNumber)
	if err == nil {
		if existingOrder.UserID == userID {
			return fmt.Errorf("order already uploaded by this user")
		} else {
			return fmt.Errorf("order already uploaded by another user")
		}
	}

	// Создаем новый заказ
	_, err = s.orderStorage.CreateOrder(userID, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (s *Service) GetUserOrders(userID int) ([]models.OrderResponse, error) {
	return s.orderStorage.GetOrdersByUserID(userID)
}

func (s *Service) GetUserBalance(userID int) (*models.BalanceResponse, error) {
	return s.balanceStorage.GetBalance(userID)
}

func (s *Service) WithdrawBalance(userID int, orderNumber string, amount float64) error {
	// Проверяем формат номера заказа
	if !s.validateOrderNumber(orderNumber) {
		return fmt.Errorf("invalid order number format")
	}

	// Проверяем достаточность средств и списываем
	err := s.balanceStorage.WithdrawBalance(userID, amount)
	if err != nil {
		return err
	}

	// Записываем операцию списания
	err = s.balanceStorage.CreateWithdrawal(userID, orderNumber, amount)
	if err != nil {
		return fmt.Errorf("failed to create withdrawal record: %w", err)
	}

	return nil
}

func (s *Service) GetUserWithdrawals(userID int) ([]models.WithdrawalResponse, error) {
	return s.balanceStorage.GetWithdrawals(userID)
}

func (s *Service) ProcessAccruals() error {
	// Получаем все заказы в обработке
	orders, err := s.orderStorage.GetProcessingOrders()
	if err != nil {
		return fmt.Errorf("failed to get processing orders: %w", err)
	}

	for _, order := range orders {
		// Запрашиваем информацию о заказе в системе расчета
		accrualInfo, err := s.accrualClient.GetOrderInfo(order.Number)
		if err != nil {
			// Если заказ не найден в системе расчета, пропускаем
			if strings.Contains(err.Error(), "order not found") {
				continue
			}
			// Если превышен лимит запросов, прерываем обработку
			if strings.Contains(err.Error(), "rate limit exceeded") {
				break
			}
			continue
		}

		// Обновляем статус заказа
		err = s.orderStorage.UpdateOrderStatus(order.Number, accrualInfo.Status, accrualInfo.Accrual)
		if err != nil {
			continue
		}

		// Если заказ обработан и есть начисление, обновляем баланс
		if accrualInfo.Status == models.OrderStatusProcessed && accrualInfo.Accrual > 0 {
			err = s.balanceStorage.UpdateBalance(order.UserID, accrualInfo.Accrual)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

// validateOrderNumber проверяет номер заказа по алгоритму Луна
func (s *Service) validateOrderNumber(number string) bool {
	// Убираем пробелы и проверяем, что это только цифры
	number = strings.ReplaceAll(number, " ", "")
	if len(number) < 2 {
		return false
	}

	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Алгоритм Луна
	sum := 0
	alternate := false

	// Проходим по цифрам справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// ValidateOrderNumber проверяет номер заказа по алгоритму Луна (публичный метод для тестирования)
func (s *Service) ValidateOrderNumber(number string) bool {
	return s.validateOrderNumber(number)
}
