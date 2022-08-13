package bot

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	uri        = "https://stooq.com/q/l/"
	queryParam = "&f=sd2t2ohlcv&h&e=csv"
	clientName = "stock-client"
)

type (
	StocksClient struct {
		Getter interface {
			Do(req *http.Request) (*http.Response, error)
		}
	}

	StockData struct {
		StockCode string
		Data      [][]string
	}

	HTTPClientError struct {
		HTTPStatusCode int
		Msg            string
		ClientName     string
		Cause          error
	}
)

func (e *HTTPClientError) Error() string {
	return fmt.Sprintf("%s requesting %s, %s: %v", http.StatusText(e.HTTPStatusCode), e.ClientName, e.Msg, e.Cause)
}

// HandleHTTPClientError is intended to handle errors happening on the client side
// as a result of a request execution.
// If the error is a Timeout, it is parsed as FailedDependency.
func HandleHTTPClientError(err error) error {
	if err, ok := errorAsTimeout(err); ok {
		return err
	}

	return err
}

func errorAsTimeout(err error) (error, bool) {
	type timeouter interface{ Timeout() bool }

	if to, ok := err.(timeouter); ok && to.Timeout() {
		return &HTTPClientError{
			HTTPStatusCode: http.StatusFailedDependency,
			Msg:            "timeout expired to request",
			Cause:          err,
		}, true
	}

	return nil, false
}

func (c StocksClient) GetStockFile(stockCode string) (StockData, error) {
	url := fmt.Sprintf("%s?s=%s%s", uri, stockCode, queryParam)
	log.Info().Msg(fmt.Sprintf("calling url %s ", url))

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Getter.Do(req)

	defer resp.Body.Close()

	switch {
	case err != nil:
		return StockData{}, HandleHTTPClientError(err)
	case resp.StatusCode != http.StatusOK:
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return StockData{}, &HTTPClientError{HTTPStatusCode: resp.StatusCode, Msg: buf.String(), ClientName: clientName}
	default:
		reader := csv.NewReader(resp.Body)
		reader.Comma = ';'
		data, err := reader.ReadAll()
		if err != nil {
			return StockData{}, err
		}
		return StockData{StockCode: stockCode, Data: data}, nil
	}
}
