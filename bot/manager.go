package bot

import (
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
	broadcasterChannelName   = "broadcast-channel"
)

type (
	stockClient interface {
		GetStockFile(stockCode string) (StockData, error)
	}
	publisher interface {
		Publish(channelName string, body []byte) error
	}
	BotMgr struct {
		stockClient stockClient
		publisher   publisher
		conn        *rabbit.Connection
	}
)

func NewBotMgr(client stockClient, publisher publisher, conn *rabbit.Connection) *BotMgr {
	return &BotMgr{
		stockClient: client,
		publisher:   publisher,
		conn:        conn,
	}
}

func (bm *BotMgr) GetAndPublishStockPrice(chatMsg chatrooms.ChatMessage) error {
	stockMsg, _ := bm.GetStockPrice(nil, getStockCode(chatMsg.Text))

	chatMsg.Text = stockMsg
	chatMsg.Username = "Bot"
	chatMsgAsByte, err := json.Marshal(chatMsg)
	if err != nil {
		log.Error().Err(err)
	}
	err = bm.publisher.Publish(broadcasterChannelName, chatMsgAsByte)
	if err != nil {
		return err
	}
	return nil
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
