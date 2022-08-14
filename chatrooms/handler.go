package chatrooms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"go-chat/api"
)

const (
	stockCommandString     = "/stock="
	messagesChannelName    = "chat-channel"
	broadcasterChannelName = "broadcast-channel"
)

var (
	clients      = make(map[*websocket.Conn]bool)
	broadcaster  = make(chan ChatMessage)
	connUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	roomId string
)

type (
	botMgr interface {
		GetStockPrice(ctx echo.Context, stockCode string) (string, *api.APIError)
	}
	publisher interface {
		Publish(channelName string, body []byte) error
		Consume(channelName string) (<-chan rabbit.Delivery, error)
	}

	Handler struct {
		BotManager  botMgr
		RedisClient *redis.Client
		Publisher   publisher
	}

	ChatMessage struct {
		Username  string `json:"username"`
		Text      string `json:"text"`
		Room      string `json:"room"`
		Timestamp string `json:"timestamp"`
	}
)

func (ch *ChatMessage) IsStockCommand() bool {
	r, _ := regexp.Compile(stockCommandString)
	return r.MatchString(ch.Text)
}

func (h *Handler) HandleConnections(c echo.Context) error {
	ws, err := connUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer ws.Close()
	clients[ws] = true
	roomId = c.Param("id")

	if h.RedisClient.Exists(roomId).Val() != 0 {
		h.sendPreviousMessages(ws)
	}

	// waiting for incoming messages
	for {
		var msg ChatMessage
		err := ws.ReadJSON(&msg)
		log.Info().Msg(fmt.Sprintf("Message received: %v\n", msg))
		if err != nil {
			delete(clients, ws)
			break
		}
		err = h.publishMessage(msg)
		if err != nil {
			log.Error().Err(err).Msg("error publishing msg")
		}
		broadcaster <- msg
	}
	return nil
}

func (h *Handler) sendPreviousMessages(ws *websocket.Conn) {
	chatMessages, err := h.RedisClient.LRange(roomId, 0, -1).Result()
	if err != nil {
		panic(err)
	}
	for _, chatMessage := range chatMessages {
		var msg ChatMessage
		json.Unmarshal([]byte(chatMessage), &msg)
		messageClient(ws, msg)
	}
}

func messageClients(msg ChatMessage) {
	for client := range clients {
		messageClient(client, msg)
	}
}

func messageClient(client *websocket.Conn, msg ChatMessage) {
	err := client.WriteJSON(msg)
	if err != nil && unsafeError(err) {
		log.Error().Err(err)
		client.Close()
		delete(clients, client)
	}
}

func (h *Handler) storeInRedis(msg ChatMessage) {
	json, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	if err := h.RedisClient.RPush(roomId, json).Err(); err != nil {
		panic(err)
	}
}

func (h *Handler) HandleMessages() {
	for {
		msg := <-broadcaster
		fmt.Printf("Message read from broadcaster: %v\n", msg)
		if !msg.IsStockCommand() {
			h.storeInRedis(msg)
			messageClients(msg)
		}
	}
}

func unsafeError(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseGoingAway) && err != io.EOF
}

func (h *Handler) publishMessage(msg ChatMessage) error {

	msgByte, _ := json.Marshal(msg)
	err := h.Publisher.Publish(messagesChannelName, msgByte)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) WaitingForQueueMsgs() {
	msgs, err := h.Publisher.Consume(broadcasterChannelName)
	if err != nil {
		log.Error().Err(err)
	}

	var stockBotChannelReceiver chan struct{}

	go func() {
		for d := range msgs {
			log.Info().Msg(fmt.Sprintf("Received message from bot: %s", d.Body))
			for client := range clients {
				var msg ChatMessage
				err := json.Unmarshal(d.Body, &msg)
				if err != nil {
					log.Error().Err(err)
				}
				err = client.WriteJSON(msg)
				if err != nil && unsafeError(err) {
					log.Error().Err(err)
					client.Close()
					delete(clients, client)
				}
			}

		}
	}()
	log.Info().Msg("Waiting for new messages in bot consumer")
	<-stockBotChannelReceiver
}
