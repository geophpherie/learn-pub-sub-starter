package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TRANSIENT = iota
	DURABLE
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(val); err != nil {
		return err
	}
	err := ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buffer.Bytes(),
		})
	if err != nil {
		return err
	}

	return nil

}

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
	handler func(T) AckType,
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

			acktype := handler(data)
			switch acktype {
			case Ack:
				fmt.Println("message ack")
				msg.Ack(false)
			case NackRequeue:
				fmt.Println("message nack requeue")
				msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("message nack discard")
				msg.Nack(false, false)
			default:
				panic("GAHH")
			}
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

	table := amqp.Table{"x-dead-letter-exchange": "peril_dlx"}

	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, table)
	if err != nil {
		return nil, amqp.Queue{}, errors.New("Unable to create queue")
	}

	channel.QueueBind(queueName, key, exchange, false, nil)

	return channel, queue, nil

}
