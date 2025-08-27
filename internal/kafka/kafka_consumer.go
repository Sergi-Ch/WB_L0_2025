package kafka

import (
	"context"
	"encoding/json"
	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"github.com/Sergi-Ch/WB_L0_2025/internal/service"
	"github.com/segmentio/kafka-go"
	"log"
)

type Consumer struct {
	reader       *kafka.Reader
	orderService *service.OrderService
}

func NewConsumer(brokers []string, topic, groupID string, orderService *service.OrderService) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &Consumer{reader: r, orderService: orderService}
}

func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Kafka consumer started...\n")
	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer context cancelled")
			return nil
		default:

			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}

				return err
			}

			var order domain.Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Printf("invalid message: %v", err)
				continue
			}

			if err := c.orderService.SaveOrder(ctx, &order); err != nil {
				log.Printf("failed to save order: %v", err)
				continue
			}

			log.Printf("order saved>>> %s", order.OrderUid)
		}
	}
}
func (c *Consumer) Close() error {
	log.Printf("closing kafka consumer...\n")
	return c.reader.Close()
}
