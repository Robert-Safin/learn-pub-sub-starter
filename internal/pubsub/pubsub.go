package pubsub

import (
	"context"
	"encoding/json"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp091.Channel, exchange, key string, val T) error {

	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        bytes,
	})
	if err != nil {
		return err
	}
	return nil
}

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

var simpleQueueTypeMap = map[SimpleQueueType]string{
	Durable:   "durable",
	Transient: "transient",
}

func DeclareAndBind(
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
) (*amqp091.Channel, amqp091.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp091.Queue{}, err
	}
	var durable bool
	var autoDelete bool
	var exclusive bool
	switch simpleQueueType {
	case int(Durable):
		durable = true
		autoDelete = false
		exclusive = false
	case int(Transient):
		durable = false
		autoDelete = true
		exclusive = true
	}

	q, err := ch.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp091.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp091.Queue{}, err
	}

	return ch, q, nil
}

func SubscribeJSON[T any](
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	ch, _, err := DeclareAndBind(conn, exchange, routing.GameLogSlug, queueName, int(Durable))
	if err != nil {
		return err
	}

	consume_ch, err := ch.Consume(queueName, "", false, false, false, false, nil)

	go func() {
		for delivery := range consume_ch {
			var data T
			json.Unmarshal(delivery.Body, &data)
			handler(data)
			delivery.Ack(false)
		}
	}()

	return nil
}
