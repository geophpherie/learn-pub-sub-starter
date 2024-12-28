package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TRANSIENT = iota
	DURABLE   = iota
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bytes,
		})
	if err != nil {
		return err
	}

	return nil

}
func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	channel, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	ch, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range ch {
			var data T
			json.Unmarshal(msg.Body, &data)
			if err != nil {
				fmt.Println("msg decode error")
			}

			handler(data)
			msg.Ack(true)
		}
	}()
	return nil
}
func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, errors.New("Unable to create channel")
	}

	var durable bool
	var autoDelete bool
	var exclusive bool
	switch simpleQueueType {
	case TRANSIENT:
		autoDelete = true
		exclusive = true
		durable = false
	case DURABLE:
		autoDelete = false
		exclusive = false
		durable = true
	default:
		return nil, amqp.Queue{}, errors.New("Unknown Queue Type")
	}

	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, errors.New("Unable to create queue")
	}

	channel.QueueBind(queueName, key, exchange, false, nil)

	return channel, queue, nil

}
