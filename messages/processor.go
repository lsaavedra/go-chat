package messages

import (
	"encoding/json"

	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"go-chat/bot"
	"go-chat/chatrooms"
)

const messagesChannelName = "chat-channel"

type (
	Processor struct {
		BotMgr      *bot.BotMgr
		MessagesMgr *MessagesMgr
		publisher   publisher
	}
	publisher interface {
		Consume(channelName string) (<-chan rabbit.Delivery, error)
	}
)

func NewProcessor(msgMgr *MessagesMgr, botMgr *bot.BotMgr, publisher publisher) *Processor {
	return &Processor{
		BotMgr:      botMgr,
		MessagesMgr: msgMgr,
		publisher:   publisher,
	}
}

func (p *Processor) WaitForQueueMsgs() {
	msgs, err := p.publisher.Consume(messagesChannelName)
	if err != nil {
		log.Error().Err(err)
	}

	var chatChannel chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Channel: %s - Received a message: %s", messagesChannelName, d.Body)
			var chatMessage chatrooms.ChatMessage
			err := json.Unmarshal(d.Body, &chatMessage)
			if err != nil {
				log.Error().Err(err).Msg("failed unmarshalling message")
			}
			if chatMessage.IsStockCommand() {
				log.Info().Msg("Calling bot manager")
				err := p.BotMgr.GetAndPublishStockPrice(chatMessage)
				if err != nil {
					log.Error().Err(err)
				}

			} else {
				log.Info().Msg("Calling msg manager")
				p.MessagesMgr.SaveMsg(chatMessage)
			}

		}
	}()
	log.Info().Msg("Waiting for new messages in message processor")
	<-chatChannel
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panic().Err(err).Msg(msg)
	}
}
