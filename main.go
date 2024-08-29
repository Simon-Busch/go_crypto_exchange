package main

import (
	"time"

	"github.com/Simon-Busch/go_crypto_exchange/client"
	"github.com/Simon-Busch/go_crypto_exchange/server"
)


func main() {
	go server.StartServer()

	time.Sleep(1 * time.Second)

	clt := client.NewClient()


	for {
		limitOrderParamsA := &client.PlaceOrderParams{
			UserID: 8,
			Bid:    false,
			Price:  10_000.0,
			Size:   500_000.0,
		}

		_, err := clt.PlaceLimitOrder(limitOrderParamsA)
		if err != nil {
			panic(err)
		}

		limitOrderParamsB := &client.PlaceOrderParams{
			UserID: 9,
			Bid:    false,
			Price:  9_000.0,
			Size:   500_000.0,
		}

		_, err = clt.PlaceLimitOrder(limitOrderParamsB)
		if err != nil {
			panic(err)
		}

		marketOrderParams := &client.PlaceOrderParams{
			UserID: 7,
			Bid:    true,
			Size:   1_000_000.0,
		}

		_, err = clt.PlaceMarketOrder(marketOrderParams)
		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
	}



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
