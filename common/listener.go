package common

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQListener struct {
	Connection *amqp.Connection
	Ch *amqp.Channel
	Queue amqp.Queue
}

/**
  the user is responsible for closing the sender
  by running RabbitMQListener.Deinit()
*/
func InitListener(routingKey string) (*RabbitMQListener, error) {
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

	queue, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(queue.Name, routingKey, ExchangeName, false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitMQListener{
		Connection: connection,
		Ch: ch,
		Queue: queue,
	}, nil
}

func (l *RabbitMQListener) Deinit() {
	l.Ch.Close()
	l.Connection.Close()
}

