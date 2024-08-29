package client

import (
	"bytes"
	"encoding/json"
	// "fmt"
	"net/http"

	"github.com/Simon-Busch/go_crypto_exchange/server"
)

const Endpoint = "http://localhost:4000"

type Client struct {
	*http.Client
}

type PlaceLimitOrderParams struct {
	UserID int64
	Bid    bool
	Price  float64
	Size   float64
}

func NewClient() *Client {
	return &Client{
		Client: http.DefaultClient,
	}
}

func (c *Client) CancelOrder() error {

	return nil
}



func (c *Client) PlaceLimitOrder(p *PlaceLimitOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserID: p.UserID,
		Type: server.LimitOrder,
		Bid: p.Bid,
		Size: p.Size,
		Price: p.Price,
		Market: server.MarketETH,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	return placeOrderResponse, nil
}
