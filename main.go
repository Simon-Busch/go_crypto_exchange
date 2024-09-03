package main

import (
	"math/rand"
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
		MinSpread: 			20, // ordersize * 2 would be good
		SeedOffset: 		40,
		ExchangeClient: clt,
		MakeInterval: 	1 * time.Second,
		PriceOffset: 		10,
	}
	maker := mm.NewMarketMaker(cfg)
	maker.Start()


	time.Sleep(1 * time.Second)
	go marketOrderPlacer(clt)

	select {}
}

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(200 * time.Millisecond)

	for {
		randomint := rand.Intn(10)
		bid := true
		if randomint > 3 { // The higher, more buying pressure -- price going up
			bid = false
		}

		order := &client.PlaceOrderParams{
			UserID: 1,
			Bid: 		bid,
			Size:		1,
		}

		_, err := c.PlaceMarketOrder(order)
		if err != nil {
			panic(err)
		}

		<- ticker.C
	}
}
