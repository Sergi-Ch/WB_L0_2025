package repository

import (
	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"sync"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]domain.Order
}

type CacheInterface interface {
	Get(orderUID string) (*domain.Order, bool)
	Set(orderUID string, order domain.Order)
}

func NewCache() *Cache {
	return &Cache{
		orders: make(map[string]domain.Order),
	}
}

func (c *Cache) Get(orderUID string) (*domain.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, ok := c.orders[orderUID]
	return &order, ok
}

func (c *Cache) Set(orderUID string, order domain.Order) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.orders[orderUID] = order
}
