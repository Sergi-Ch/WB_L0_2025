package service

import (
	"context"
	"errors"
	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"github.com/Sergi-Ch/WB_L0_2025/internal/repository"
	"log"
)

type OrderServiceInterface interface {
	SaveOrder(ctx context.Context, order *domain.Order) error
	GetOrderByID(ctx context.Context, id string) (*domain.Order, error)
}

type OrderService struct {
	cache    repository.CacheInterface
	postgres repository.PostgresRepInterface
}

func NewOrderService(pg repository.PostgresRepInterface, cache repository.CacheInterface) *OrderService {
	return &OrderService{
		postgres: pg,
		cache:    cache,
	}
}

func (s *OrderService) SaveOrder(ctx context.Context, order *domain.Order) error {
	if err := s.postgres.SaveOrders(ctx, order); err != nil {
		log.Printf("failed to save order in postgres: %v", err)
		return err
	}

	s.cache.Set(order.OrderUid, *order)

	return nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {

	if order, ok := s.cache.Get(id); ok {
		return order, nil
	}

	order, err := s.postgres.GetByID(ctx, id)
	if err != nil {
		log.Printf("order not found in postgres: %v", err)
		return nil, err
	}

	if order != nil {
		s.cache.Set(order.OrderUid, *order)
		return order, nil
	}

	return nil, errors.New("order not found")
}
