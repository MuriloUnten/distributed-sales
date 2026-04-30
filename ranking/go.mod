module github.com/MuriloUnten/distributed-sales/ranking

go 1.26.1

require (
	github.com/MuriloUnten/distributed-sales/common v0.0.0-20260430223603-ea8a3accbcd0
	github.com/rabbitmq/amqp091-go v1.11.0
)

replace github.com/MuriloUnten/distributed-sales/common => ../common
