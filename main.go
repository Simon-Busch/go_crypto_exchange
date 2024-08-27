package main

import (
	// "context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"net/http"
	"strconv"

	"github.com/Simon-Busch/go_crypto_exchange/orderbook"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)


const (
	// Dummy anvil priv key
	exchangePrivKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	MarketOrder OrderType = "MARKET"
	LimitOrder 	OrderType = "LIMIT"

	MarketETH Market = "ETH"
)

type (
	OrderType string
	Market string

	PlaceOrderRequest struct {
		Type 		OrderType // Limit or market
		Bid 		bool
		Size 		float64
		Price 	float64
		Market 	Market
		UserID 	int64
	}

	Order struct {
		ID 				int64
		Price 		float64
		Size 			float64
		Bid 			bool
		Timestamp int64
	}

	OrderbookData struct {
		TotalBidVolume 	float64
		TotalAskVolume	float64
		Asks 						[]*Order
		Bids 						[]*Order
	}

	Exchange struct {
		Client 					*ethclient.Client
		PrivateKey 			*ecdsa.PrivateKey // Exchange hot wallet
		orderbooks			map[Market]*orderbook.Orderbook
		orders 					map[int64]int64
		Users 					map[int64]*User
	}

	MatchedOrder struct {
		Size 			float64
		Price 		float64
		ID 				int64
	}

	User struct {
		ID 						int64
		PrivateKey 		*ecdsa.PrivateKey
	}
)


func main() {
	e := echo.New()

	e.HTTPErrorHandler = httpErrorHandler

	// RPC address -- in this case anvil
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(exchangePrivKey, client)
	if err != nil {
		log.Fatal(err)
	}

	// Add a user
	pk := "2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"
	user := NewUser(pk)
	user.ID = 11
	ex.Users[user.ID] = user

	fmt.Printf("User: %+v\n", user)


	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.handleCancelOrder)

	// ctx := context.Background()
	// // Dummy address from anvil
	// // address := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	// // balance, err := client.BalanceAt(ctx, address, nil)
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// // fmt.Printf("before balance: %s",balance)

	// // Associated anvil private key
	// //NB : remove 0x from the private key
	// privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// publicKey := privateKey.Public()
	// publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	// if !ok {
	// 	log.Fatal("error casting public key to ECDSA")
	// }

	// fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// value := big.NewInt(1000000000000000000) // in wei (1 eth)
	// gasLimit := uint64(21000) // in units
	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// toAddress := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	// tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	// fmt.Printf("tx ==> %+v\n", tx)

	// chainID, err := client.NetworkID(context.Background()) // 31337 for localhost / Anvil
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("tx sent ===> %s", signedTx.Hash().Hex())


	// if err := client.SendTransaction(context.Background(), signedTx); err != nil {
	// 	log.Fatal(err)
	// }

	// toBalance, err := client.BalanceAt(ctx, toAddress, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf(" after balance: %s\n",toBalance)

	// fromBalance, err := client.BalanceAt(ctx, fromAddress, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf(" from address after balance: %s\n",fromBalance)

	e.Start(":4000")
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
	code := http.StatusInternalServerError
	msg := "Internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message.(string)
	}

	c.JSON(code, map[string]any{"msg": msg})
}

func NewUser(privateKey string) *User {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic("Impossible to create new user")
	}

	return &User{
		PrivateKey: pk,
	}
}

func NewExchange(privateKey string, client *ethclient.Client) (*Exchange, error) {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	privKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	return &Exchange{
		Client: 			client,
		PrivateKey: 	privKey,
		orderbooks: 	orderbooks,
		Users: 				make(map[int64]*User),
		orders: 			make(map[int64]int64),
	}, nil
}


func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "market not found"})
	}

	orderbookData := OrderbookData{
		TotalBidVolume: 	ob.BidTotalVolume(),
		TotalAskVolume: 	ob.AskTotalVolume(),
		Asks: 						[]*Order{},
		Bids: 						[]*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				ID: 				order.ID,
				Price: 			limit.Price,
				Size: 			order.Size,
				Bid: 				order.Bid,
				Timestamp: 	order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				ID: 				order.ID,
				Price: 			limit.Price,
				Size: 			order.Size,
				Bid: 				order.Bid,
				Timestamp: 	order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	idStr := c.Param("id") // param are always string
	id, err := strconv.ParseInt(idStr, 10, 64)
	if (err != nil) {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "invalid id"})
	}

	ob := ex.orderbooks[MarketETH] // Get rid of hardcoding
	order := ob.Orders[id]
	ob.CancelOrder(order)

	return c.JSON(http.StatusOK, map[string]any{"msg": "order cancelled", "id": id})
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchesOrders := make([]*MatchedOrder, len(matches))

	isBid := order.Bid

	for i := 0 ; i < len(matches); i++ {
		id := matches[i].Bid.ID
		if isBid {
			id = matches[i].Ask.ID
		}
		match := matches[i]
		matchesOrders[i] = &MatchedOrder{
			Price: 			match.Price,
			Size: 			match.SizeFilled,
			ID: 				id,
		}
	}

	return matches, matchesOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	user, ok := ex.Users[order.UserID]

	if !ok {
		return fmt.Errorf("user not found")
	}

	fmt.Printf("User in handlePlaceLimitOrder: %+v\n", user)

	exchangePubKey := ex.PrivateKey.Public()
	publicKeyECDSA, ok := exchangePubKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting exchange public key to ECDSA")
	}

	toAddr := crypto.PubkeyToAddress(*publicKeyECDSA)
	amount := big.NewInt(int64(order.Size * price))

	// Transfer  user => exchange wallet
	return transferETH(ex.Client, user.PrivateKey, toAddr, amount)
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}

	market := Market(placeOrderData.Market)
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	if placeOrderData.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"msg": "error placing limit"})
		}
		return c.JSON(http.StatusOK, map[string]any{"msg": "limit order placed"})
	}

	if placeOrderData.Type == MarketOrder {
		matches, matchesOrders := ex.handlePlaceMarketOrder(market, order)

		if err := ex.handleMatches(matches); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]any{"matches": matchesOrders})
	}

	return c.JSON(http.StatusBadRequest, map[string]any{"msg": "invalid order type"})
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	// for _, match := range matches {

	// }
	return nil
}
