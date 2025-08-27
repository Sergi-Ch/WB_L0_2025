package main

import (
	"context"
	"fmt"
	"github.com/Sergi-Ch/WB_L0_2025/internal/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	prHttp "github.com/Sergi-Ch/WB_L0_2025/internal/delivery/http"
	"github.com/Sergi-Ch/WB_L0_2025/internal/repository"
	"github.com/Sergi-Ch/WB_L0_2025/internal/service"
)

func runMigrations(dsn string) error {
	ctx := context.Background()

	// Используем pgxpool вместо database/sql
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Читаем файл миграции
	migrationSQL, err := os.ReadFile("migrations/001_init.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Выполняем миграцию
	_, err = pool.Exec(ctx, string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	log.Println("Migrations executed successfully")
	return nil
}
func main() {
	// подгрузка переменных окружения
	//err := godotenv.Load(".env")
	//if err != nil {
	//	log.Fatalf("error of loading .env %v", err)
	//}
	password := os.Getenv("DATABASE_PASSWORD")
	dataBaseName := os.Getenv("DATABASE_NAME")
	userName := os.Getenv("USER_NAME")
	port := os.Getenv("APP_PORT")
	dsn := "postgres://" + userName + ":" + password + "@postgres:5432/" + dataBaseName
	pgRepo, err := repository.NewPostgresRepository(dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	log.Println("Running database migrations...")
	if err := runMigrations(dsn); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	//инициализация слоев
	cache := repository.NewCache()
	orderService := service.NewOrderService(pgRepo, cache)

	handler := prHttp.NewOrderHandler(orderService)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//подключение kafka
	brokers := []string{"kafka:29092"}
	topic := "orders"
	groupID := "order-service"

	consumer := kafka.NewConsumer(brokers, topic, groupID, orderService)
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatalf("kafka consumer failed: %v", err)
		}
	}()

	go func() {
		//запуск сервера
		log.Printf("Server started on :%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	//graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down...")
	cancel() //кафка

	ctxTimeOut, cancelTimeOut := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelTimeOut()
	if err := srv.Shutdown(ctxTimeOut); err != nil {
		log.Printf("http server shutdown error: %v", err)
	}

	if err := consumer.Close(); err != nil {
		log.Printf("Error closing kafka consumer>>> %v", err)
	}
	log.Println("Server stopped gracefully")
}
