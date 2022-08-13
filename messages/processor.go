package messages

import (
	"encoding/json"
	"regexp"
	"strings"

	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"go-chat/bot"
	"go-chat/chatrooms"
)

type (
	Processor struct {
		// tendrá los managers de cada uno
		// cliente de rabbit se suscribe y consume el canal
		BotMgr      *bot.BotMgr
		MessagesMgr *MessagesMgr
		Conn        *rabbit.Connection
	}
)

func NewProcessor(msgMgr *MessagesMgr, botMgr *bot.BotMgr, conn *rabbit.Connection) *Processor {
	return &Processor{
		BotMgr:      botMgr,
		MessagesMgr: msgMgr,
		Conn:        conn,
	}
}

func (p *Processor) ReadAndProcess() {
	ch, err := p.Conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"chat-channel", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Channel: %s - Received a message: %s", q.Name, d.Body)
			var chatMessage chatrooms.ChatMessage
			err := json.Unmarshal(d.Body, &chatMessage)
			if err != nil {
				// ojo con esto porque n o debería poder pasar
				log.Error().Err(err).Msg("failed unmarshalling message")
			}
			if isStockMessage(chatMessage.Text) {
				log.Info().Msg("Calling bot manager")
				err := p.BotMgr.GetAndPublishStockMessage(chatMessage)
				if err != nil {
					log.Error().Err(err)
				}

			} else {
				log.Info().Msg("Calling msg manager")
				p.MessagesMgr.CreateFromEvent(chatMessage)
			}

		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panic().Err(err).Msg(msg)
	}
}

func isStockMessage(message string) bool {
	r, _ := regexp.Compile("/stock=")
	return r.MatchString(message)
}

func getStockCode(message string) string {
	// probably here it could handle not understood messages or format
	return strings.Split(message, "=")[1]
}
