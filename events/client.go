package events

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

var channel *rabbit.Channel

type EventMetadata struct {
	ID      uuid.UUID
	From    string
	Source  string
	Channel string
	Subject []byte
}

type (
	QueueClient struct {
		serviceName string
		conn        *rabbit.Connection
	}
)

func NewQueueClient(serviceName string, conn *rabbit.Connection) *QueueClient {
	return &QueueClient{
		serviceName: serviceName,
		conn:        conn,
	}
}

func (qc *QueueClient) connect(queueName string) *rabbit.Channel {
	// create a channel
	ch, err := qc.conn.Channel()
	failOnError(err, "Failed to open a channel")

	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	channel = ch
	return ch
}

func (qc *QueueClient) CloseConnection() {
	defer channel.Close()
	log.Info().Msg("closing rabbit connection")
}

func (qc *QueueClient) Publish(channelName string, body []byte) error {
	ch := qc.connect(channelName)

	err := ch.PublishWithContext(
		context.TODO(),
		"",          // exchange
		channelName, // routing key
		false,       // mandatory
		false,       // immediate
		rabbit.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		log.Error().Err(err).Msg("Failed to publish msg")
		return err
	}
	log.Info().Msg(fmt.Sprintf(" [x]-> Publisher: %s - Channel: %s - Body %s\n", qc.serviceName, channelName, string(body)))
	return nil
}

func (qc *QueueClient) Consume(channelName string) (<-chan rabbit.Delivery, error) {
	ch := qc.connect(channelName)

	msgs, err := ch.Consume(
		channelName, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	log.Info().Msg(fmt.Sprintf(" <-[x] Consumer: %s - Channel: %s\n", qc.serviceName, channelName))
	return msgs, err
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panic().Err(err).Msg(msg)
	}
}
