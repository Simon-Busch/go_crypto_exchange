package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Simon-Busch/go_crypto_exchange/client"
	"github.com/Simon-Busch/go_crypto_exchange/server"
)

const (
	maxOrders = 3
)

var (
	tick = 2 * time.Second
)

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(4 * time.Second)

	for {
		// Sell
		marketSellOrder := &client.PlaceOrderParams{
			UserID: 7,
			Bid:    false,
			Size:   1000.0,
		}

		orderResp, err := c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderID)
		}


		// Buy
		marketBuyOrder := &client.PlaceOrderParams{
			UserID: 7,
			Bid:    true,
			Size:   800.0,
		}

		marketOrderResp, err := c.PlaceMarketOrder(marketBuyOrder)
		if err != nil {
			log.Println(marketOrderResp.OrderID)
		}

		<- ticker.C
	}
}


func makeMarketSimple(clt *client.Client) {
	ticker := time.NewTicker(tick)

	for {

		_, err := clt.GetOrders(8)
		if err != nil {
			log.Println(err)
		}

		// fmt.Printf("===================================\n")
		// fmt.Printf("Orders for user [8] => %+v\n", orders)
		// fmt.Printf("===================================\n")

		bestAsk, err := clt.GetBestAsk()
		if err != nil {
			log.Println(err)
		}
		bestBid, err := clt.GetBestBid()
		if err != nil {
			log.Println(err)
		}

		spread := math.Abs(bestBid - bestAsk)
		fmt.Println("Exchange Spread => ", spread)

		// Place the bids
		if len(myBids) < maxOrders {
			bidLimit := &client.PlaceOrderParams{
				UserID: 8,
				Bid:    true,
				Price:  bestBid + 100,
				Size:   1_000,
			}

			bidOrderResp, err := clt.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(bidOrderResp.OrderID)
			}
			myBids[bidLimit.Price] = bidOrderResp.OrderID
		}

		// Place the asks
		if len(myAsks) < maxOrders {
			askLimit := &client.PlaceOrderParams{
				UserID: 8,
				Bid:    false,
				Price:  bestAsk - 100,
				Size:   1_000,
			}

			askOrderResp, err := clt.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(askOrderResp.OrderID)
			}
			myAsks[askLimit.Price] = askOrderResp.OrderID
		}


		fmt.Println("Best ask => ", bestAsk)
		fmt.Println("Best bid => ", bestBid)

		<- ticker.C // more performant than time.Sleep
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

	go makeMarketSimple(clt)

	time.Sleep(1 * time.Second)
	marketOrderPlacer(clt)

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
