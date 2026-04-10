package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/MuriloUnten/distributed-sales/common"
)

type RabbitMQListener struct {
	connection *amqp.Connection
	ch *amqp.Channel
	queue amqp.Queue
}

/**
  the user is responsible for closing the sender
  by running RabbitMQListener.Deinit()
*/
func InitListener() (*RabbitMQListener, error) {
	connection, err := amqp.Dial(common.Url)
	if err != nil {
		return nil, err
	}

	ch, err := connection.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		common.ExchangeName,
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(queue.Name, common.PublishedKey , common.ExchangeName, false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitMQListener{
		connection: connection,
		ch: ch,
		queue: queue,
	}, nil
}

func (l *RabbitMQListener) Deinit() {
	l.ch.Close()
	l.connection.Close()
}

