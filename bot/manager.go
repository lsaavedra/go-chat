package bot

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"

	"go-chat/api"
)

const (
	stockMessage           = "%v quote is $%v per share"
	stockNotFoundedMessage = "Invalid stock code for command /stock=%v"
	noDataIdentifier       = "N/D"
)

type (
	stockClient interface {
		GetStockFile(stockCode string) (StockData, error)
	}
	BotMgr struct {
		stockClient stockClient
	}
)

func NewBotMgr(client stockClient) *BotMgr {
	return &BotMgr{
		stockClient: client,
	}
}

func (bm *BotMgr) GetStockPrice(ctx echo.Context, stockCode string) (string, *api.APIError) {
	stockData, err := bm.stockClient.GetStockFile(stockCode)
	if err != nil {
		return "", &api.APIError{HTTPStatusCode: http.StatusBadRequest, Cause: err}
	}

	return readFileAndGetStockPrice(stockData), nil
}

func readFileAndGetStockPrice(stockData StockData) string {
	stockValues := strings.Split(stockData.Data[1][0], ",")
	currentStockPrice := stockValues[3]
	if stockValues[3] == noDataIdentifier {
		return fmt.Sprintf(stockNotFoundedMessage, stockData.StockCode)
	}
	return fmt.Sprintf(stockMessage, strings.ToUpper(stockData.StockCode), currentStockPrice)
}
