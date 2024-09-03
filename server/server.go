package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"sync"

	"net/http"
	"strconv"

	"github.com/Simon-Busch/go_crypto_exchange/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

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
		UserID 		int64
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
		mu 							sync.RWMutex
		PrivateKey 			*ecdsa.PrivateKey // Exchange hot wallet
		orderbooks			map[Market]*orderbook.Orderbook
		Orders 					map[int64][]*orderbook.Order // map users to his orders
		Users 					map[int64]*User
	}

	MatchedOrder struct {
		UserID 		int64
		Size 			float64
		Price 		float64
		ID 				int64
	}

	User struct {
		ID 						int64
		PrivateKey 		*ecdsa.PrivateKey
	}

	APIError struct {
		Error string
	}
)


func StartServer() {
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

	// Add a user 9
	pk9 := "2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"
	// Add a user 8
	pk8 := "dbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97"
	// Add 1
	pk1 := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	ex.registerUser(pk8, 8)
	ex.registerUser(pk9, 9)
	ex.registerUser(pk1, 1)
	// address1 := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	// balance1, err := client.BalanceAt(context.Background(), common.HexToAddress(address1), nil)
	// fmt.Printf("User 1- starting balance: %s\n", balance1)

	e.POST("/order", ex.handlePlaceOrder)

	e.GET("/trades/:market", ex.HandleGetTrades)
	e.GET("/order/:userID", ex.handleGetOrders)
	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/bestbid", ex.handleGetBestBid)
	e.GET("/book/:market/bestask", ex.handleGetBestAsk)

	e.DELETE("/order/:id", ex.handleCancelOrder)

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

func NewUser(privateKey string, id int64) *User {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic("Impossible to create new user")
	}

	return &User{
		ID: 				id,
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
		Orders: 			make(map[int64][]*orderbook.Order),
		mu: 					sync.RWMutex{},
	}, nil
}

type GetOrdersResponse struct {
	Asks []Order
	Bids []Order
}


func (ex *Exchange) registerUser(pk string, userID int64) {
	user := NewUser(pk, userID)
	ex.Users[userID] = user

	logrus.WithFields(logrus.Fields{
		"userID": userID,
		"address": crypto.PubkeyToAddress(user.PrivateKey.PublicKey),
	}).Info("New exchange user")
}

func (ex *Exchange) HandleGetTrades(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, APIError{Error:"market not found"})
	}

	return c.JSON(http.StatusOK, ob.Trades)
}

func (ex *Exchange) handleGetOrders(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return err
	}

	ex.mu.RLock()
	orderbookOrders := ex.Orders[int64(userID)]
	orderResp := &GetOrdersResponse{
		Asks: []Order{},
		Bids: []Order{},
	}


	for i := 0; i < len(orderbookOrders); i++ {
		// it could be that the order is getting filled even though it is included in this response
		// We need to double check if the Limit is not nil
		if orderbookOrders[i].Limit == nil {
			continue
		}
		order := Order{
			ID:    			orderbookOrders[i].ID,
			UserID: 		orderbookOrders[i].UserID,
			Price:     	orderbookOrders[i].Limit.Price,
			Size:      	orderbookOrders[i].Size,
			Timestamp: 	orderbookOrders[i].Timestamp,
			Bid:       	orderbookOrders[i].Bid,
		}

		if order.Bid {
			orderResp.Bids = append(orderResp.Bids, order)
		} else {
			orderResp.Asks = append(orderResp.Asks, order)
		}
	}
	ex.mu.RUnlock()

	return c.JSON(http.StatusOK, orderResp)
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
				UserID: 		order.UserID,
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
				UserID: 		order.UserID,
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


type PriceResponse struct {
	Price float64
}

func (ex *Exchange) handleGetBestBid(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]

	if len(ob.Bids()) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "no bids available"})
	}

	bestBidPrice := ob.Bids()[0].Price // They are already sorted so we know this is the best

	pr := &PriceResponse{
		Price: bestBidPrice,
	}

	return c.JSON(http.StatusOK, pr)
}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]

	if len(ob.Asks()) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "no asks available"})
	}

	bestAskResponse := ob.Asks()[0].Price // They are already sorted so we know this is the best

	pr := &PriceResponse{
		Price: bestAskResponse,
	}

	return c.JSON(http.StatusOK, pr)
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

	log.Println("order cancelled => id: ", id)

	return c.JSON(http.StatusOK, map[string]any{"msg": "order cancelled", "id": id})
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchesOrders := make([]*MatchedOrder, len(matches))

	isBid := order.Bid

	totalSizeFilled := 0.0
	sumPrice := 0.0
	for i := 0 ; i < len(matches); i++ {
		var (
			match = matches[i]
			limitUserID = match.Bid.UserID
			id = match.Bid.ID
		)

		if isBid {
			limitUserID = match.Ask.UserID
			id = match.Ask.ID
		}

		matchesOrders[i] = &MatchedOrder{
			UserID: 		limitUserID,
			Price: 			match.Price,
			Size: 			match.SizeFilled,
			ID: 				id,
		}

		totalSizeFilled += match.SizeFilled
		sumPrice += match.Price
	}

	avgPrice := sumPrice / float64(len(matches))

	logrus.WithFields(logrus.Fields{
		"type": 		order.Type(),
		"size": 		totalSizeFilled,
		"avgPrice": avgPrice,
	}).Info("Filled market order")

	newOrderMap := make(map[int64][]*orderbook.Order)

	ex.mu.Lock()
	for userID, orderbookOrders := range ex.Orders {
		for i := 0 ; i < len(orderbookOrders); i++ {
			// If the order is not filled we place in the map copy
			// -> the size of the order is 0
			if !orderbookOrders[i].IsFilled() {
				newOrderMap[userID] = append(newOrderMap[userID], orderbookOrders[i])
			}
		}
	}

	ex.Orders = newOrderMap
	ex.mu.Unlock()

	return matches, matchesOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	// Keep track of the user's orders
	ex.mu.Lock()
	ex.Orders[order.UserID] = append(ex.Orders[order.UserID], order)
	ex.mu.Unlock()

	// log.Printf("new LIMIT order => type: [%t] || price[%.2f] || size [%.2f] || userID [%d]", order.Bid, order.Limit.Price, order.Size, order.UserID)

	return nil
}

type PlaceOrderResponse struct {
	OrderID int64
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}

	market := Market(placeOrderData.Market)
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	// Limit orders
	if placeOrderData.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"msg": "error placing limit"})
		}

		resp := &PlaceOrderResponse{
			OrderID: order.ID,
		}

		return c.JSON(http.StatusOK, resp)
	}

	// Market orders
	if placeOrderData.Type == MarketOrder {
		matches, _ := ex.handlePlaceMarketOrder(market, order)

		if err := ex.handleMatches(matches); err != nil {
			return err
		}


		resp := &PlaceOrderResponse{
			OrderID: order.ID,
		}

		return c.JSON(http.StatusOK, resp)
	}


	return c.JSON(http.StatusBadRequest, map[string]any{"msg": "invalid order type"})
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserID]
		if !ok {
			return fmt.Errorf("user not found for ask: %d", match.Ask.UserID)
		}

		toUser, ok := ex.Users[match.Bid.UserID]
		if !ok {
			return fmt.Errorf("user not found for bid: %d", match.Bid.UserID)
		}

		toAddress := crypto.PubkeyToAddress(toUser.PrivateKey.PublicKey)


		// Only needed for the fees later on
		// exchangePubKey := ex.PrivateKey.Public()
		// publicKeyECDSA, ok := exchangePubKey.(*ecdsa.PublicKey)
		// if !ok {
		// 	return fmt.Errorf("error casting exchange public key to ECDSA")
		// }

		amount := big.NewInt(int64(match.SizeFilled))
		transferETH(ex.Client, fromUser.PrivateKey, toAddress, amount)


		// CheckBalances
		// publicKey := fromUser.PrivateKey.Public()
		// publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		// if !ok {
		// 	return fmt.Errorf("error casting public key to ECDSA")
		// }

		// fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
		// fromBalance, err := ex.Client.BalanceAt(context.Background(), fromAddress, nil)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Printf(" from address after balance: %s\n",fromBalance)

		// toBalance, err := ex.Client.BalanceAt(context.Background(), toAddress, nil)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Printf(" to address after balance: %s\n",toBalance)

	}
	return nil
}

func transferETH(client *ethclient.Client, fromPrivKey *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()
	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}


	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(ctx) // 31337 for localhost / Anvil
	if err != nil {
		log.Fatal(err)
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		log.Fatal(err)
	}


	return client.SendTransaction(ctx, signedTx)
}
