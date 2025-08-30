// scripts/performance_benchmark.go
package main

import (
	"context"
	"fmt"
	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"github.com/Sergi-Ch/WB_L0_2025/internal/service"
	"log"
	"sync"
	"time"
)

type MockRedis struct {
	orders map[string]*domain.Order
	mu     sync.RWMutex // для потокобезопасности
}

func NewMockRedis() *MockRedis {
	return &MockRedis{
		orders: make(map[string]*domain.Order),
	}
}

func (m *MockRedis) Get(orderUID string) (*domain.Order, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	order, exists := m.orders[orderUID]
	return order, exists
}

func (m *MockRedis) Set(orderUID string, order domain.Order) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[orderUID] = &order
}

type MockPostgres struct {
	orders map[string]*domain.Order
	mu     sync.RWMutex
}

func NewMockPostgres() *MockPostgres {
	return &MockPostgres{
		orders: make(map[string]*domain.Order),
	}
}

func (m *MockPostgres) SaveOrders(ctx context.Context, order *domain.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := 0; i < 1000000; i++ {
		_ = i * i
	}

	m.orders[order.OrderUid] = order
	return nil
}

func (m *MockPostgres) GetByID(ctx context.Context, orderUID string) (*domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := 0; i < 1000000; i++ {
		_ = i * i
	}

	order, exists := m.orders[orderUID]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func main() {
	fmt.Println("start")

	// Создаем тестовый заказ
	now := time.Now()
	order := &domain.Order{
		OrderUid:    "perf-test-order-123",
		TrackNumber: "PERF123456",
		Delivery: domain.Delivery{
			Name:  "Test User",
			Phone: "+79991112233",
			Email: "test@example.com",
		},
		Payment: domain.Payment{
			Transaction: "PERF_TXN_123456",
			Amount:      2500,
		},
		Items: []domain.Item{
			{
				ChrtId: 10001,
				Name:   "Test Product",
				Price:  1000,
			},
		},
		DateCreated: now,
	}

	redisCache := NewMockRedis()
	postgres := NewMockPostgres()

	service := service.NewOrderService(postgres, redisCache)
	ctx := context.Background()

	fmt.Println("Сохраняем тестовый заказ...")
	if err := service.SaveOrder(ctx, order); err != nil {
		log.Fatal("Ошибка сохранения заказа:", err)
	}

	fmt.Println("\nТест 1: Получение из кеша (100 запросов) ")
	start := time.Now()

	for i := 0; i < 100; i++ {
		_, err := service.GetOrderByID(ctx, order.OrderUid)
		if err != nil {
			log.Printf("Ошибка: %v", err)
		}
	}

	cacheTime := time.Since(start)
	fmt.Printf("Время выполнения 100 запросов к кешу: %v\n", cacheTime)
	fmt.Printf("Среднее время одного запроса: %v\n", cacheTime/100)

	fmt.Println("\nТест 2: Получение из БД (100 запросов) ")
	redisCache.orders = make(map[string]*domain.Order)

	start = time.Now()

	for i := 0; i < 100; i++ {
		_, err := service.GetOrderByID(ctx, order.OrderUid)
		if err != nil {
			log.Printf("Ошибка: %v", err)
		}
	}

	dbTime := time.Since(start)
	fmt.Printf("Время выполнения 100 запросов к БД: %v\n", dbTime)
	fmt.Printf("Среднее время одного запроса: %v\n", dbTime/100)

	fmt.Println("\n РЕЗУЛЬТАТЫ")
	fmt.Printf("Память (кеш):  %v всего (%v среднее)\n", cacheTime, cacheTime/100)
	fmt.Printf("Вычисления (БД): %v всего (%v среднее)\n", dbTime, dbTime/100)

	if cacheTime > 0 {
		ratio := float64(dbTime) / float64(cacheTime)
		fmt.Printf("Кеш быстрее БД в %.1f раз\n", ratio)
	}

	fmt.Printf("Разница: %v\n", dbTime-cacheTime)

}
