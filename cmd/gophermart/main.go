package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/draco777/gophermart/internal/accrual"
	"github.com/draco777/gophermart/internal/config"
	"github.com/draco777/gophermart/internal/handlers"
	"github.com/draco777/gophermart/internal/middleware"
	"github.com/draco777/gophermart/internal/service"
	"github.com/draco777/gophermart/internal/storage"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Запускаем миграции
	if err := storage.RunMigrations(cfg.DatabaseURI); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Подключаемся к базе данных
	st, err := storage.New(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer st.Close()

	// Создаем хранилища
	userStorage := storage.NewUserStorage(st.DB())
	orderStorage := storage.NewOrderStorage(st.DB())
	balanceStorage := storage.NewBalanceStorage(st.DB())

	// Создаем клиент для системы расчета баллов
	accrualClient := accrual.NewAccrualClient(cfg.AccrualSystemAddress)

	// Создаем сервис
	svc := service.NewService(userStorage, orderStorage, balanceStorage, accrualClient)

	// Создаем handlers
	h := handlers.NewHandlers(svc)

	// Настраиваем роутинг
	mux := http.NewServeMux()

	// Тестовый эндпоинт
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Gophermart API is running"))
	})

	// Публичные эндпоинты
	mux.HandleFunc("/api/user/register", h.Register)
	mux.HandleFunc("/api/user/login", h.Login)

	// Защищенные эндпоинты с middleware аутентификации
	mux.Handle("/api/user/orders", middleware.AuthMiddleware(http.HandlerFunc(h.OrdersHandler)))
	mux.Handle("/api/user/balance", middleware.AuthMiddleware(http.HandlerFunc(h.BalanceHandler)))
	mux.Handle("/api/user/balance/withdraw", middleware.AuthMiddleware(http.HandlerFunc(h.WithdrawBalance)))
	mux.Handle("/api/user/withdrawals", middleware.AuthMiddleware(http.HandlerFunc(h.GetWithdrawals)))

	// Создаем сервер
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: mux,
	}

	// Запускаем горутину для обработки начислений
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := svc.ProcessAccruals(); err != nil {
				log.Printf("Failed to process accruals: %v", err)
			}
		}
	}()

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server starting on %s", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
