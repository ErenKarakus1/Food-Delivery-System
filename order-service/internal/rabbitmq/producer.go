package rabbitmq

import (
	"encoding/json"
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/model"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch    *amqp091.Channel
	queue amqp091.Queue
	conn  *amqp091.Connection
}

func NewPublisher(rabbitmqURL string) (*Publisher, error) {
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
	return &Publisher{
		ch:    ch,
		queue: q,
		conn:  conn,
	}, nil
}

func (p *Publisher) PublishOrderReady(orderID uuid.UUID) error {
	event := model.OrderReadyForPickupEvent{OrderID: orderID.String()}
	payload, err := json.Marshal(event)
	if err != nil {
		return errors.New("internal server error")
	}
	err = p.ch.Publish(
		"",
		p.queue.Name,
		false,
		false,
		amqp091.Publishing{
			Body: payload,
		},
	)
	if err != nil {
		return errors.New("internal server error")
	}
	return nil
}

func (p *Publisher) Close() error {
	var err1, err2 error
	if p.ch != nil {
		err1 = p.ch.Close()
	}
	if p.conn != nil {
		err2 = p.conn.Close()
	}
	if err1 != nil || err2 != nil {
		return errors.New("internal server error")
	}
	return nil
}
