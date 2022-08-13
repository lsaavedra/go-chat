package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"go-chat/api"
	"go-chat/chatrooms"
)

const (
	stockMessage             = "%v quote is $%v per share"
	stockNotFoundedMessage   = "Invalid stock code for command /stock=%v"
	stockServiceNotAvailable = "Stock service is not available"
	noDataIdentifier         = "N/D"
)

type (
	stockClient interface {
		GetStockFile(stockCode string) (StockData, error)
	}
	BotMgr struct {
		stockClient stockClient
		Conn        *rabbit.Connection
	}
)

func NewBotMgr(client stockClient, conn *rabbit.Connection) *BotMgr {
	return &BotMgr{
		stockClient: client,
		Conn:        conn,
	}
}

func (bm *BotMgr) GetAndPublishStockMessage(chatMsg chatrooms.ChatMessage) error {
	// create a channel
	ch, err := bm.Conn.Channel()
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

	stockMsg, err := bm.GetStockPrice(nil, getStockCode(chatMsg.Text))
	// improve this log error because is returning *ApiError{nil}
	//if !errors.Is(err, &api.APIError{}) {
	//	log.Info().Err(err)
	//}
	chatMsg.Text = stockMsg
	chatMsg.Username = "Bot"
	chatMsgAsByte, err := json.Marshal(chatMsg)
	if err != nil {
		log.Error().Err(err)
	}

	err = ch.PublishWithContext(
		context.TODO(),
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		rabbit.Publishing{
			ContentType: "text/plain",
			Body:        chatMsgAsByte,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Channel: %s - Sent %s\n", q.Name, string(chatMsgAsByte))
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panic().Err(err).Msg(msg)
	}
}

func (bm *BotMgr) GetStockPrice(ctx echo.Context, stockCode string) (string, *api.APIError) {
	stockData, err := bm.stockClient.GetStockFile(stockCode)
	if err != nil {
		return stockServiceNotAvailable, &api.APIError{HTTPStatusCode: http.StatusBadRequest, Cause: err}
	}

	return readFileAndGetStockPrice(stockData), &api.APIError{}
}

func readFileAndGetStockPrice(stockData StockData) string {
	stockValues := strings.Split(stockData.Data[1][0], ",")
	currentStockPrice := stockValues[3]
	if stockValues[3] == noDataIdentifier {
		return fmt.Sprintf(stockNotFoundedMessage, stockData.StockCode)
	}
	return fmt.Sprintf(stockMessage, strings.ToUpper(stockData.StockCode), currentStockPrice)
}

func getStockCode(message string) string {
	// probably here it could handle not understood messages or format
	return strings.Split(message, "=")[1]
}
