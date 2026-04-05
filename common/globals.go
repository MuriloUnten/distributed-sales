package common

const (
	Url = "amqp://guest:guest@localhost:5672/assignment1"

	ExchangeName = "promocoes"
	ReceivedKey = "promocao.recebida"
	PublishedKey = "promocao.publicada"
	
	LogsExchangeName = "logs"
	InfoKey = "info"
	WarningKey = "warn"
	ErrorKey = "error"
	DebugKey = "debug"
)
