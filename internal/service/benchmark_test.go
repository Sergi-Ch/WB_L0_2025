package service

import (
	"context"
	"errors"
	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"testing"
	"time"
)

func BenchmarkGetOrderFromCache(b *testing.B) {

	order := createTestOrder()

	cache := &MockCache{orders: make(map[string]*domain.Order)}
	postgres := &MockPostgres{orders: make(map[string]*domain.Order)}

	service := NewOrderService(postgres, cache)

	ctx := context.Background()
	if err := service.SaveOrder(ctx, order); err != nil {
		b.Fatalf("failed to save order: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GetOrderByID(ctx, order.OrderUid)
			if err != nil {
				b.Errorf("failed to get order from cache: %v", err)
			}
		}
	})
}

func BenchmarkGetOrderFromDB(b *testing.B) {

	order := createTestOrder()

	cache := &MockCache{orders: make(map[string]*domain.Order)}
	postgres := &MockPostgres{orders: make(map[string]*domain.Order)}

	service := NewOrderService(postgres, cache)

	ctx := context.Background()
	if err := service.SaveOrder(ctx, order); err != nil {
		b.Fatalf("failed to save order: %v", err)
	}

	cache.Delete(order.OrderUid)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GetOrderByID(ctx, order.OrderUid)
			if err != nil {
				b.Errorf("failed to get order from DB: %v", err)
			}
		}
	})
}

func createTestOrder() *domain.Order {
	now := time.Now()
	return &domain.Order{
		OrderUid:          "test-order-benchmark-123",
		TrackNumber:       "TRACK123456",
		Entry:             "ENTRY123",
		InternalSignature: "INTERNAL_SIG_123",
		CustomerId:        "CUSTOMER_123",
		DeliveryService:   "DELIVERY_SERVICE_123",
		Shardkey:          "SHARD_123",
		SmId:              12345,
		DateCreated:       now,
		OofShard:          "OOF_SHARD_123",
		Delivery: domain.Delivery{
			Name:    "John Doe",
			Phone:   "+79991234567",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Lenina st, 123",
			Region:  "Moscow region",
			Email:   "john@example.com",
		},
		Payment: domain.Payment{
			Transaction:  "TXN123456",
			RequestId:    "REQ123456",
			Currency:     "RUB",
			Provider:     "SBER",
			Amount:       1500,
			PaymentDt:    int(now.Unix()),
			Bank:         "SBERBANK",
			DeliveryCost: 500,
			GoodsTotal:   1000,
			CustomFee:    0,
		},
		Items: []domain.Item{
			{
				ChrtId:      12345,
				TrackNumber: "TRACK123456",
				Price:       500,
				Rid:         "RID123456",
				Name:        "Product 1",
				Sale:        10,
				Size:        "M",
				TotalPrice:  900,
				NmId:        67890,
				Brand:       "Brand 1",
				Status:      1,
			},
			{
				ChrtId:      12346,
				TrackNumber: "TRACK123456",
				Price:       600,
				Rid:         "RID123457",
				Name:        "Product 2",
				Sale:        0,
				Size:        "L",
				TotalPrice:  600,
				NmId:        67891,
				Brand:       "Brand 2",
				Status:      1,
			},
		},
		Locale: "ru",
	}
}

type MockCache struct {
	orders map[string]*domain.Order
}

func (m *MockCache) Set(key string, order domain.Order) {
	m.orders[key] = &order
}

func (m *MockCache) Get(key string) (*domain.Order, bool) {
	order, exists := m.orders[key]
	return order, exists
}

func (m *MockCache) Delete(key string) {
	delete(m.orders, key)
}

type MockPostgres struct {
	orders map[string]*domain.Order
}

func (m *MockPostgres) SaveOrders(ctx context.Context, order *domain.Order) error {
	time.Sleep(10 * time.Millisecond)
	m.orders[order.OrderUid] = order
	return nil
}

func (m *MockPostgres) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	time.Sleep(10 * time.Millisecond)
	order, exists := m.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}
