package repository

import (
	"context"
	"fmt"
	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

type PostgresRepInterface interface {
	SaveOrders(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, orderUID string) (*domain.Order, error)
}

func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) SaveOrders(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `insert into orders (order_uid,track_number, entry, locale, internal_signature, 
                    customer_id, delivery_service, shardkey, sm_id,date_created, oof_shard)
                    values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, order.OrderUid, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerId, order.DeliveryService,
		order.Shardkey, order.SmId, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("insert orders faiked: %w", err)
	}

	_, err = tx.Exec(ctx, `insert into deliveries (order_uid, name, phone, zip, city, address, region, email) values 
                                                                                    ($1,$2,$3,$4,$5,$6,$7,$8)
	`, order.OrderUid, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("insert delivery failed: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (
			order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, order.OrderUid, order.Payment.Transaction, order.Payment.RequestId,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("insert payment failed: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, order.OrderUid, item.ChrtId, item.TrackNumber, item.Price, item.Rid,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("insert item failed: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, orderUID string) (*domain.Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
			o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,

			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,

			p.transaction, p.request_id, p.currency, p.provider,
			p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee,

			i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size,
			i.total_price, i.nm_id, i.brand, i.status
		FROM orders o
		LEFT JOIN deliveries d ON d.order_uid = o.order_uid
		LEFT JOIN payments p  ON p.order_uid = o.order_uid
		LEFT JOIN items i    ON i.order_uid = o.order_uid
		WHERE o.order_uid = $1;
	`, orderUID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var order domain.Order
	order.Items = []domain.Item{}

	firstRow := true
	for rows.Next() {
		var item domain.Item

		if firstRow {

			err = rows.Scan(
				&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
				&order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId, &order.DateCreated, &order.OofShard,

				&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
				&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,

				&order.Payment.Transaction, &order.Payment.RequestId, &order.Payment.Currency, &order.Payment.Provider,
				&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
				&order.Payment.GoodsTotal, &order.Payment.CustomFee,

				&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale,
				&item.Size, &item.TotalPrice, &item.NmId, &item.Brand, &item.Status,
			)
			if err != nil {
				return nil, fmt.Errorf("scan first row failed: %w", err)
			}
			firstRow = false
		} else {

			err = rows.Scan(
				new(string), new(string), new(string), new(string), new(string),
				new(string), new(string), new(string), new(int), new(time.Time), new(string),

				new(string), new(string), new(string), new(string), new(string), new(string), new(string), new(string),

				new(string), new(string), new(string), new(string),
				new(int), new(int64), new(string), new(int), new(int), new(int),

				&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale,
				&item.Size, &item.TotalPrice, &item.NmId, &item.Brand, &item.Status,
			)
			if err != nil {
				return nil, fmt.Errorf("scan item row failed: %w", err)
			}
		}

		order.Items = append(order.Items, item)
	}

	if firstRow {
		return nil, fmt.Errorf("order not found")
	}

	return &order, nil
}

func (r *PostgresRepository) Close() error {
	if r.db != nil {
		r.db.Close()
	}
	return nil
}
