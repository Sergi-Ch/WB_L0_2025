package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

type RedisInterface interface {
	Get(orderUID string) (*domain.Order, bool)
	Set(orderUID string, order domain.Order)
}

func NewRedisCache(addr string, ttl time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return &RedisCache{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisCache) Get(orderUID string) (*domain.Order, bool) {
	ctx := context.Background()
	data, err := r.client.Get(ctx, "order:"+orderUID).Bytes()
	if err != nil {
		return nil, false
	}

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, false
	}

	return &order, true
}

func (r *RedisCache) Set(orderUID string, order domain.Order) {
	ctx := context.Background()
	data, _ := json.Marshal(order)
	r.client.Set(ctx, "order:"+orderUID, data, r.ttl)
}

func (r *RedisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
