module github.com/MuriloUnten/distributed-sales/gateway

go 1.26.1

require (
	github.com/MuriloUnten/distributed-sales/common v1.0.0
	github.com/rabbitmq/amqp091-go v1.10.0
)

replace github.com/MuriloUnten/distributed-sales/common => ../common
