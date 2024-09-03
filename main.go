package main

import (
	"time"

	"github.com/Simon-Busch/go_crypto_exchange/client"
	"github.com/Simon-Busch/go_crypto_exchange/mm"
	"github.com/Simon-Busch/go_crypto_exchange/server"
)


func main() {
	go server.StartServer()
	time.Sleep(1 * time.Second)

	clt := client.NewClient()

	cfg := mm.Config{
		UserID: 				9,
		OrderSize: 			10,
		MinSpread: 			100,
		SeedOffset: 		400,
		ExchangeClient: clt,
		MakeInterval: 	1 * time.Second,
	}
	maker := mm.NewMarketMaker(cfg)
	maker.Start()

	// go makeMarketSimple(clt)

	time.Sleep(1 * time.Second)
	go marketOrderPlacer(clt)

	select {}
}

// func makeMarketSimple(c *client.Client) {
// 	ticker := time.NewTicker(1 * time.Second)

// 	for {
// 		bestAsk, err := c.GetBestAsk()
// 		if err != nil {
// 			panic(err)
// 		}
// 		bestBid, err := c.GetBestBid()
// 		if err != nil {
// 			panic(err)
// 		}

// 		if bestAsk == 0.0 && bestBid == 0.0 {
// 			seedMarket(c)
// 			continue
// 		}

// 		fmt.Printf("Best ask: %.2f\n", bestAsk)
// 		fmt.Printf("Best bid: %.2f\n", bestBid)

// 		<- ticker.C
// 	}
// }

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		buyOrder := &client.PlaceOrderParams{
			UserID: 1,
			Bid: 		true,
			Size:		1,
		}

		_, err := c.PlaceMarketOrder(buyOrder)
		if err != nil {
			panic(err)
		}

		sellOrder := &client.PlaceOrderParams{
			UserID: 1,
			Bid: 		false,
			Size:		1,
		}

		_, err = c.PlaceMarketOrder(sellOrder)
		if err != nil {
			panic(err)
		}

		<- ticker.C
	}
}


// const ethPrice = 2158.0
// func seedMarket(c *client.Client) {
// 	currentPrice := ethPrice // Should be an async call to get the current price
// 	priceOffset := 100.0

// 	bidOrder := &client.PlaceOrderParams{
// 		UserID: 9,
// 		Bid: 		true,
// 		Price: 	currentPrice - priceOffset,
// 		Size:		10,
// 	}
// 	_, err := c.PlaceLimitOrder(bidOrder)
// 	if err != nil {
// 		panic(err)
// 	}

// 	askOrder := &client.PlaceOrderParams{
// 		UserID: 9,
// 		Bid: 		false,
// 		Price: 	currentPrice + priceOffset,
// 		Size:		10,
// 	}
// 	_, err = c.PlaceLimitOrder(askOrder)
// 	if err != nil {
// 		panic(err)
// 	}
// }
