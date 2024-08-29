package main

import (
	"fmt"
	"time"

	"github.com/Simon-Busch/go_crypto_exchange/client"
	"github.com/Simon-Busch/go_crypto_exchange/server"
)


func main() {
	go server.StartServer()

	time.Sleep(1 * time.Second)

	clt := client.NewClient()

	// Place a bid limit order
	bidParams := &client.PlaceLimitOrderParams{
		UserID: 8,
		Bid:    true,
		Price:  10_000.0,
		Size:   1_000.0,
	}

	go func() {
		for  {
			resp, err := clt.PlaceLimitOrder(bidParams)

			if err != nil {
				panic(err)
			}

			fmt.Println("order id => ", resp.OrderID)

			time.Sleep(1 * time.Second)
		}
	}()


	// Place an ask limit order
	askParams := &client.PlaceLimitOrderParams{
		UserID: 8,
		Bid:    false,
		Price:  8_000.0,
		Size:   1_000.0,
	}

	for  {
		resp, err := clt.PlaceLimitOrder(askParams)

		if err != nil {
			panic(err)
		}

		fmt.Println("order id => ", resp.OrderID)
		time.Sleep(1 * time.Second)
	}

	select {}
}
