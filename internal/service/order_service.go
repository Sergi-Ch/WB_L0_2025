package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"github.com/Sergi-Ch/WB_L0_2025/internal/repository"
	"log"
	"regexp"
	"strings"
	"time"
)

type OrderServiceInterface interface {
	SaveOrder(ctx context.Context, order *domain.Order) error
	GetOrderByID(ctx context.Context, id string) (*domain.Order, error)
}

type OrderService struct {
	cache    repository.CacheInterface
	postgres repository.PostgresRepInterface
}

func NewOrderService(pg repository.PostgresRepInterface, redis repository.RedisInterface) *OrderService {
	return &OrderService{
		postgres: pg,
		cache:    redis,
	}
}

func (s *OrderService) SaveOrder(ctx context.Context, order *domain.Order) error {

	if err := s.validateOrder(order); err != nil {
		log.Printf("order validation failed: %v", err)
		return fmt.Errorf("invalid order: %w", err)
	}

	if err := s.postgres.SaveOrders(ctx, order); err != nil {
		log.Printf("failed to save order in postgres: %v", err)
		return err
	}

	s.cache.Set(order.OrderUid, *order)

	return nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {

	if id == "" {
		return nil, errors.New("order id is required")
	}

	// Проверка длины ID
	if len(id) > 50 {
		return nil, errors.New("order id is too long")
	}

	// Проверка на допустимые символы
	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validID.MatchString(id) {
		return nil, errors.New("order id contains invalid characters")
	}

	if order, ok := s.cache.Get(id); ok {
		return order, nil
	}

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

func (s *OrderService) validateOrder(order *domain.Order) error {
	if order == nil {
		return errors.New("order is nil")
	}

	// Валидация OrderUid
	if order.OrderUid == "" {
		return errors.New("order_uid is required")
	}
	if len(order.OrderUid) > 50 {
		return errors.New("order_uid is too long (max 50 characters)")
	}
	if strings.Contains(order.OrderUid, " ") {
		return errors.New("order_uid cannot contain spaces")
	}

	// Валидация TrackNumber
	if order.TrackNumber == "" {
		return errors.New("track_number is required")
	}
	if len(order.TrackNumber) > 50 {
		return errors.New("track_number is too long (max 50 characters)")
	}

	// Валидация даты
	if order.DateCreated.IsZero() {
		return errors.New("date_created is required")
	}
	if order.DateCreated.After(time.Now().Add(time.Hour)) {
		return errors.New("date_created cannot be in the future")
	}

	// Валидация Delivery
	if err := s.validateDelivery(&order.Delivery); err != nil {
		return fmt.Errorf("delivery validation failed: %w", err)
	}

	// Валидация Payment
	if err := s.validatePayment(&order.Payment); err != nil {
		return fmt.Errorf("payment validation failed: %w", err)
	}

	// Валидация Items
	if len(order.Items) == 0 {
		return errors.New("at least one item is required")
	}
	for i, item := range order.Items {
		if err := s.validateItem(&item); err != nil {
			return fmt.Errorf("item[%d] validation failed: %w", i, err)
		}
	}

	return nil
}

func (s *OrderService) validateDelivery(delivery *domain.Delivery) error {
	if delivery == nil {
		return errors.New("delivery is nil")
	}

	if delivery.Name == "" {
		return errors.New("delivery name is required")
	}
	if len(delivery.Name) > 100 {
		return errors.New("delivery name is too long (max 100 characters)")
	}

	if delivery.Phone == "" {
		return errors.New("delivery phone is required")
	}

	if delivery.Email == "" {
		return errors.New("delivery email is required")
	}
	if len(delivery.Email) > 100 {
		return errors.New("delivery email is too long (max 100 characters)")
	}
	if !strings.Contains(delivery.Email, "@") || !strings.Contains(delivery.Email, ".") {
		return errors.New("delivery email format is invalid")
	}

	return nil
}

func (s *OrderService) validatePayment(payment *domain.Payment) error {
	if payment == nil {
		return errors.New("payment is nil")
	}

	if payment.Transaction == "" {
		return errors.New("payment transaction is required")
	}
	if len(payment.Transaction) > 50 {
		return errors.New("payment transaction is too long (max 50 characters)")
	}

	if payment.Amount <= 0 {
		return errors.New("payment amount must be greater than 0")
	}
	if payment.Amount > 1000000000 { // 10 миллионов
		return errors.New("payment amount is too large")
	}

	if payment.Currency == "" {
		return errors.New("payment currency is required")
	}
	if len(payment.Currency) > 3 {
		return errors.New("payment currency code is invalid (max 3 characters)")
	}

	return nil
}

func (s *OrderService) validateItem(item *domain.Item) error {
	if item == nil {
		return errors.New("item is nil")
	}

	if item.Name == "" {
		return errors.New("item name is required")
	}
	if len(item.Name) > 200 {
		return errors.New("item name is too long (max 200 characters)")
	}

	if item.Price <= 0 {
		return errors.New("item price must be greater than 0")
	}
	if item.Price > 100000000 { // 100 миллионов
		return errors.New("item price is too large")
	}

	if item.TotalPrice < 0 {
		return errors.New("item total_price cannot be negative")
	}

	if item.ChrtId <= 0 {
		return errors.New("item chrt_id must be greater than 0")
	}

	return nil
}
