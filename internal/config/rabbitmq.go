package config

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitMQ *amqp.Connection
var RabbitChannel *amqp.Channel

func InitRabbitMQ(url string) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal("failed connect rabbitmq:", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("failed create channel:", err)
	}

	err = ch.ExchangeDeclare(
		"ticket_events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal("failed declare exchange:", err)
	}

	RabbitMQ = conn
	RabbitChannel = ch

	log.Println("RabbitMQ connected")
}
