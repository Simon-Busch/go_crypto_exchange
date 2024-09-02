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
			Size:   100.0,
		}

		orderResp, err := c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderID)
		}

		otherMarketSellOrder := &client.PlaceOrderParams{
			UserID: 9,
			Bid:    false,
			Size:   100.0,
		}

		orderResp, err = c.PlaceMarketOrder(otherMarketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderID)
		}

		// Buy
		marketBuyOrder := &client.PlaceOrderParams{
			UserID: 9,
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

const userID = 8

/*
Basics of market making:
- Market making is a strategy where a trader simultaneously places both buy and sell orders in an attempt to profit from the bid-ask spread.
- The bid-ask spread is the difference between the highest price that a buyer is willing to pay for an asset and the lowest price that a seller is willing to accept.
- real goal here is not the spread for lower market maker but the volume of trades.
- Market makers provide liquidity to the market by placing orders on both sides of the order book, which helps ensure that there are always buyers and sellers available to trade.
- Market making can be a profitable strategy if done correctly, but it also carries risks, such as the risk of losses if the market moves against the trader.

*/

// Liquidity provider
func makeMarketSimple(clt *client.Client) {
	ticker := time.NewTicker(tick)

	for {

		orders, err := clt.GetOrders(userID)
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
		if len(orders.Bids) < maxOrders {
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
		}

		// Place the asks
		if len(orders.Asks) < maxOrders {
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
		Size:   2_000.0,
	}
	_, err := c.PlaceLimitOrder(ask)
	if err != nil {
		return err
	}


	bid := &client.PlaceOrderParams{
		UserID: 8,
		Bid:    true,
		Price:  9_000.0,
		Size:   2_000.0,
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

	select {}
}
