package chatrooms

import (
	"context"
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
	stockCommandString = "/stock="
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

	Handler struct {
		BotManager  botMgr
		QueueCon    *rabbit.Connection
		RedisClient *redis.Client
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
		fmt.Printf("Message received: %v\n", msg)
		if err != nil {
			delete(clients, ws)
			break
		}
		fmt.Printf("Active clients: %v\n", clients)
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
		//panic(err)
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
		log.Printf("error: %v", err)
		client.Close()
		delete(clients, client)
	}
}

func (h *Handler) storeInRedis(msg ChatMessage) {
	json, err := json.Marshal(msg)
	if err != nil {
		//panic(err)
	}

	if err := h.RedisClient.RPush(roomId, json).Err(); err != nil {
		//panic(err)
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

// If a message is sent while a client is closing, ignore the error
func unsafeError(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseGoingAway) && err != io.EOF
}

func (h *Handler) publishMessage(msg ChatMessage) error {
	// create a channel
	ch, err := h.QueueCon.Channel()
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

	msgByte, _ := json.Marshal(msg)

	err = ch.PublishWithContext(
		context.TODO(),
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		rabbit.Publishing{
			ContentType: "text/plain",
			Body:        msgByte,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Channel: %s - Sent %s\n", q.Name, msgByte)
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panic().Err(err).Msg(msg)
	}
}

func (h *Handler) ReadAndProcess() {
	ch, err := h.QueueCon.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"broadcast-channel", // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
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

	var stockBotChannelReceiver chan struct{}

	go func() {
		for d := range msgs {
			// for this case we should only send the message for the client who sends the message
			log.Printf("Received a message: %s", d.Body)
			for client := range clients {
				var msg ChatMessage
				err := json.Unmarshal(d.Body, &msg)
				if err != nil {
					log.Error().Err(err)
				}
				err = client.WriteJSON(msg)
				if err != nil && unsafeError(err) {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}

		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-stockBotChannelReceiver
}
