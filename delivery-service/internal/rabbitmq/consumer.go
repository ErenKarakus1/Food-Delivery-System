package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	pool   *pgxpool.Pool
	conn   *amqp091.Connection
	ch     *amqp091.Channel
	queue  amqp091.Queue
	client *http.Client
}

func NewConsumer(rabbitmqURL string, pool *pgxpool.Pool) (*Consumer, error) {
	conn, err := amqp091.Dial(rabbitmqURL)
	if err != nil {
		return nil, errors.New("internal server error")
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, errors.New("internal server error")
	}
	q, err := ch.QueueDeclare(
		"order.ready_for_pickup",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, errors.New("internal server error")
	}
	return &Consumer{
		pool:   pool,
		conn:   conn,
		ch:     ch,
		queue:  q,
		client: &http.Client{Timeout: time.Second * 5},
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume(
		c.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.New("internal server error")
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			var event model.OrderReadyForPickupEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				msg.Nack(false, false)
				continue
			}
			if err := c.handleOrderReady(ctx, event); err != nil {
				if errors.Is(err, repository.ErrCourierNotFound) {
					msg.Nack(false, true)
				} else {
					msg.Nack(false, false)
				}
				continue
			}
			msg.Ack(false)
		}
	}
}

func (c *Consumer) handleOrderReady(ctx context.Context, event model.OrderReadyForPickupEvent) error {
	orderID, err := uuid.Parse(event.OrderID)
	if err != nil {
		return errors.New("invalid order id")
	}
	delivery, err := repository.GetDeliveryByOrderID(ctx, c.pool, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrDeliveryNotFound) {
			delivery = service.CreateDelivery(orderID)
			if err := repository.CreateDelivery(ctx, c.pool, delivery); err != nil {
				return errors.New("internal server error")
			}
		} else {
			return errors.New("internal server error")
		}
	}
	if delivery.CourierID != uuid.Nil {
		return nil
	}
	courier, err := repository.GetAvailableCourierForDelivery(ctx, c.pool, delivery.ID)
	if err != nil {
		if errors.Is(err, repository.ErrCourierNotFound) {
			return repository.ErrCourierNotFound
		}
		return errors.New("internal server error")
	}
	if err := repository.AssignDelivery(ctx, c.pool, delivery.ID, courier.ID); err != nil {
		return errors.New("internal server error")
	}
	if err := repository.SetUnavailable(ctx, c.pool, courier.ID); err != nil {
		return errors.New("internal server error")
	}
	url := fmt.Sprintf("http://localhost:8082/orders/courier/%s/delivery_created", orderID.String())
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return errors.New("internal server error")
	}
	req.Header.Set("X-User-Role", "courier")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.New("order service error")
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("order service error")
	}
	return nil
}

func (c *Consumer) Close() error {
	var err1, err2 error
	if c.ch != nil {
		err1 = c.ch.Close()
	}
	if c.conn != nil {
		err2 = c.conn.Close()
	}
	if err1 != nil || err2 != nil {
		return errors.New("internal server error")
	}
	return nil
}
