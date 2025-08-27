package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"github.com/segmentio/kafka-go"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	brokers := []string{"kafka:29092"}
	topic := "orders"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	numOrders := rand.Intn(11) + 5 // 5–15 заказов
	log.Printf("Sending %d orders...\n", numOrders)

	for i := 0; i < numOrders; i++ {
		order := generateRandomOrder()
		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("failed to marshal order: %v", err)
			continue
		}

		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(order.OrderUid),
				Value: data,
			},
		)
		if err != nil {
			log.Printf("failed to send order: %v", err)
			continue
		}

		fmt.Printf("✅ Order sent: %s (%d items)\n", order.OrderUid, len(order.Items))
		time.Sleep(300 * time.Millisecond)
	}
}

func generateRandomOrder() domain.Order {
	uid := fmt.Sprintf("test-%d", rand.Intn(1000000))
	numItems := rand.Intn(5) + 1 // 1–5 товаров

	items := make([]domain.Item, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = domain.Item{
			ChrtId:      rand.Intn(100000),
			TrackNumber: fmt.Sprintf("TRACK-%d", rand.Intn(9999)),
			Price:       rand.Intn(1000),
			Rid:         fmt.Sprintf("rid-%d", rand.Intn(100000)),
			Name:        fmt.Sprintf("Product-%d", i+1),
			Sale:        rand.Intn(50),
			Size:        "M",
			TotalPrice:  rand.Intn(1000),
			NmId:        rand.Intn(10000),
			Brand:       "BrandX",
			Status:      202,
		}
	}

	return domain.Order{
		OrderUid:    uid,
		TrackNumber: fmt.Sprintf("TRACK-%d", rand.Intn(9999)),
		Entry:       "WBIL",
		Locale:      "en",
		Delivery: domain.Delivery{
			Name:    "John Doe",
			Phone:   "+123456789",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Lenina 1",
			Region:  "Moscow",
			Email:   "john@example.com",
		},
		Payment: domain.Payment{
			Transaction:  uid,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       rand.Intn(5000),
			DeliveryCost: 1500,
			GoodsTotal:   1000,
			CustomFee:    0,
		},
		Items:           items,
		CustomerId:      "test-customer",
		DeliveryService: "meest",
		Shardkey:        "9",
		SmId:            99,
		DateCreated:     time.Now(),
		OofShard:        "1",
	}
}
