package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/MuriloUnten/distributed-sales/common"
)


type RabbitMQSender struct {
	connection *amqp.Connection
	ch *amqp.Channel
}

/**
  the user is responsible for closing the sender
  by running RabbitMQListener.Deinit()
*/
func InitSender() (*RabbitMQSender, error) {
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

	return &RabbitMQSender{
		connection: connection,
		ch: ch,
	}, nil
}

func (s *RabbitMQSender) Deinit() {
	s.ch.Close()
	s.connection.Close()
}
