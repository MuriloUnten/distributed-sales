module github.com/MuriloUnten/distributed-sales/notification

go 1.26.1

require (
	github.com/MuriloUnten/distributed-sales/common v0.0.0-20260430233350-340ffc4100ee
	github.com/rabbitmq/amqp091-go v1.11.0
)

replace github.com/MuriloUnten/distributed-sales/common => ../common
