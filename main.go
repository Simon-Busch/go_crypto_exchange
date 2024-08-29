package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Simon-Busch/go_crypto_exchange/client"
	"github.com/Simon-Busch/go_crypto_exchange/server"
)

var tick = 2 * time.Second


func makeMarketSimple(client *client.Client) {
	ticker := time.NewTicker(tick)

	for {
		<- ticker.C // more performant than time.Sleep

		bestAsk, err := client.GetBestAsk()
		if err != nil {
			log.Println(err)
		}
		bestBid, err := client.GetBestBid()
		if err != nil {
			log.Println(err)
		}

		spread := math.Abs(bestBid - bestAsk)


		fmt.Println("Exchange Spread => ", spread)
		fmt.Println("Best ask => ", bestAsk)
		fmt.Println("Best bid => ", bestBid)
	}
}

func seedMarket(c *client.Client) error {
	ask := &client.PlaceOrderParams{
		UserID: 8,
		Bid:    false,
		Price:  10_000.0,
		Size:   1_000_000.0,
	}
	_, err := c.PlaceLimitOrder(ask)
	if err != nil {
		return err
	}


	bid := &client.PlaceOrderParams{
		UserID: 8,
		Bid:    true,
		Price:  9_000.0,
		Size:   1_000_000.0,
	}
	_, err = c.PlaceLimitOrder(bid)
	if err != nil {
		return err
	}

	return nil
}


func main() {
	go server.StartServer()

	time.Sleep(1 * time.Second)

	clt := client.NewClient()

	if err := seedMarket(clt); err != nil {
		panic(err)
	}

	makeMarketSimple(clt)

	// for {
	// 	limitOrderParamsA := &client.PlaceOrderParams{
	// 		UserID: 8,
	// 		Bid:    false,
	// 		Price:  10_000.0,
	// 		Size:   5_000_000.0,
	// 	}

	// 	_, err := clt.PlaceLimitOrder(limitOrderParamsA)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	limitOrderParamsB := &client.PlaceOrderParams{
	// 		UserID: 9,
	// 		Bid:    false,
	// 		Price:  9_000.0,
	// 		Size:   500_000.0,
	// 	}

	// 	_, err = clt.PlaceLimitOrder(limitOrderParamsB)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	buyLimitOrder := &client.PlaceOrderParams{
	// 		UserID: 7,
	// 		Bid:    true,
	// 		Price:  11_000.0,
	// 		Size:   500_000.0,
	// 	}
	// 	_, err = clt.PlaceLimitOrder(buyLimitOrder)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	marketOrderParams := &client.PlaceOrderParams{
	// 		UserID: 7,
	// 		Bid:    true,
	// 		Size:   1_000_000.0,
	// 	}

	// 	_, err = clt.PlaceMarketOrder(marketOrderParams)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// bestBidPrice, err := clt.GetBestBid()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("best bid price => ", bestBidPrice)

	// bestAskPrice, err := clt.GetBestAsk()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("best ask price => ", bestAskPrice)

	// 	time.Sleep(1 * time.Second)
	// }



	/********************************
	* Place limit order bid and ask *
	*********************************/
	// Place a bid limit order
	// bidParams := &client.PlaceOrderParams{
	// 	UserID: 8,
	// 	Bid:    true,
	// 	Price:  10_000.0,
	// 	Size:   1_000.0,
	// }

	// go func() {
	// 	for  {
	// 		resp, err := clt.PlaceLimitOrder(bidParams)

	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		fmt.Println("order id => ", resp.OrderID)

	// 		if err := clt.CancelOrder(resp.OrderID); err != nil {
	// 			panic(err)
	// 		}

	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	// // Place an ask limit order
	// askParams := &client.PlaceOrderParams{
	// 	UserID: 8,
	// 	Bid:    false,
	// 	Price:  8_000.0,
	// 	Size:   1_000.0,
	// }

	// for  {
	// 	resp, err := clt.PlaceLimitOrder(askParams)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Println("order id => ", resp.OrderID)
	// 	time.Sleep(1 * time.Second)
	// }

	select {}
}
