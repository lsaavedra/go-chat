package chatrooms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"go-chat/api"
)

var (
	clients     = make(map[*websocket.Conn]bool) // is a list of all the currently active clients (or open WebSockets).
	broadcaster = make(chan ChatMessage)         //  is a single channel that is responsible for sending and receiving our ChatMessage data structure.
	upgrader    = websocket.Upgrader{            // is a bit of a clunker; it’s necessary to “upgrade” Gorilla’s incoming requests into a WebSocket connection.
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type (
	botMgr interface {
		GetStockPrice(ctx echo.Context, stockCode string) (string, *api.APIError)
	}

	Handler struct {
		BotManager botMgr
		QueueCon   *rabbit.Connection
	}

	ChatMessage struct {
		Username  string `json:"username"`
		Text      string `json:"text"`
		Room      string `json:"room"`
		Timestamp string `json:"timestamp"`
	}
)

func (ch *ChatMessage) isStockCommand() bool {
	r, _ := regexp.Compile("/stock=")
	return r.MatchString(ch.Text)
}

func (h *Handler) HandleConnections(c echo.Context) error {
	/*
		When a new user joins the chat, three things should happen:
		1. They should be set up to receive messages from other clients.
		2. They should be able to send their own messages.
		3. They should receive a full history of the previous chat (backed by Redis).
	*/
	// resolving point 1.
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Fatal().Err(err)
	}
	// ensure connection close when function returns
	defer ws.Close()
	clients[ws] = true
	fmt.Printf("Client connections %v\n", clients)

	// resolving point 2. // waiting for incoming messages
	for {
		var msg ChatMessage
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		fmt.Printf("Message received: %v\n", msg)
		if err != nil {
			delete(clients, ws)
			break
		}
		// send new message to the channel
		//broadcaster <- msg
		//
		fmt.Printf("Active clients: %v\n", clients)
		if msg.isStockCommand() {
			err = h.publishMessage(msg)
			if err != nil {
				log.Error().Err(err).Msg("error publishing msg")
			}
		} else {
			broadcaster <- msg
		}
		// ----------------- publish message to queueu
		/*

			err = h.publishMessage(msg)
			if err != nil {
				log.Error().Err(err).Msg("error publishing msg")
			}
		*/
		// ------------------------------------------
	}
	return nil
}

func (h *Handler) HandleMessages() {
	for {
		// grab any next message from channel
		msg := <-broadcaster
		fmt.Printf("Message read from broadcaster: %v\n", msg)
		/// validate is stock message
		/*
			if isStockMessage(msg.Text) {
				stockCode := getStockCode(msg.Text)
				fmt.Println("stock message required", stockCode)
				result, _ := h.BotManager.GetStockPrice(nil, stockCode)
				fmt.Println("stock price", result)
				msg.Text = result
			}
		*/
		/// end validate
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil && unsafeError(err) {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// If a message is sent while a client is closing, ignore the error
func unsafeError(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseGoingAway) && err != io.EOF
}

func isStockMessage(message string) bool {
	r, _ := regexp.Compile("/stock=")
	return r.MatchString(message)
}

func getStockCode(message string) string {
	// probably here it could handle not understood messages or format
	return strings.Split(message, "=")[1]
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

	var forever chan struct{}

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
	<-forever
}
