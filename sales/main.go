package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	URL = "amqp://guest:guest@localhost:5672/"
)

func main() {
	connection, err := amqp.Dial(URL)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	ch, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"promocoes",
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
}
