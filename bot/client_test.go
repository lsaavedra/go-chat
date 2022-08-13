package bot

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type (
	client struct {
		fail       bool
		statusCode int
	}
)

func (t client) Do(req *http.Request) (*http.Response, error) {
	response := &http.Response{Status: "200", StatusCode: t.statusCode, Body: io.NopCloser(strings.NewReader("any body"))}
	if t.fail {
		return response, errors.New("failed to do request")
	}
	return response, nil
}

func NewClientMock(fail bool, statusCode int) *client {
	return &client{
		fail:       fail,
		statusCode: statusCode,
	}
}

func TestStocksClient_GetStockFile(t *testing.T) {
	tests := []struct {
		name        string
		stockClient *StocksClient
		stockCode   string
		wantErr     bool
	}{
		{
			name:        "Get stock - Success",
			stockClient: &StocksClient{NewClientMock(false, 200)},
			stockCode:   "aapl.us",
			wantErr:     false,
		},
		{
			name:        "Get stock - Error in request",
			stockClient: &StocksClient{NewClientMock(true, 200)},
			stockCode:   "aapl.us",
			wantErr:     true,
		},
		{
			name:        "Get stock - Error from server response",
			stockClient: &StocksClient{NewClientMock(false, 500)},
			stockCode:   "aapl.us",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.stockClient.GetStockFile(tt.stockCode); (err != nil) != tt.wantErr {
				t.Errorf("GetVerification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
