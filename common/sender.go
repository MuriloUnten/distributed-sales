package common

import (
	amqp "github.com/rabbitmq/amqp091-go"
)


type RabbitMQSender struct {
	Connection *amqp.Connection
	Ch *amqp.Channel
}

/**
  the user is responsible for closing the sender
  by running RabbitMQListener.Deinit()
*/
func InitSender() (*RabbitMQSender, error) {
	connection, err := amqp.Dial(Url)
	if err != nil {
		return nil, err
	}

	ch, err := connection.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		ExchangeName,
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
		Connection: connection,
		Ch: ch,
	}, nil
}

func (s *RabbitMQSender) Deinit() {
	s.Ch.Close()
	s.Connection.Close()
}
