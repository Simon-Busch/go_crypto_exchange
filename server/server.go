package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"net/http"
	"strconv"

	"github.com/Simon-Busch/go_crypto_exchange/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	user9 := NewUser(pk9, 9)
	ex.Users[user9.ID] = user9
	addressStr9 := "0xa0Ee7A142d267C1f36714E4a8F75612F20a79720"
	balance9, err := client.BalanceAt(context.Background(), common.HexToAddress(addressStr9), nil)
	fmt.Printf("User 9 - buyer - starting balance: %s\n", balance9)



	// Add a user 8
	pk8 := "dbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97"
	user8 := NewUser(pk8, 8)
	ex.Users[user8.ID] = user8
	fmt.Printf("User 8: %+v\n", user8)
	addressStr8 := "0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f"
	balance8, err := client.BalanceAt(context.Background(), common.HexToAddress(addressStr8), nil)
	fmt.Printf("User 8  - seller - starting balance: %s\n", balance8)


	// Add Bob
	bobPK := "4bbbf85ce3377467afe5d46f804f221813b2bb87f24d81f60f1fcdbf7cbf4356"
	userBob := NewUser(bobPK, 7)
	ex.Users[userBob.ID] = userBob
	fmt.Printf("User Bob: %+v\n", userBob)

	addressBob := "0x14dC79964da2C08b23698B3D3cc7Ca32193d9955"
	balanceBob, err := client.BalanceAt(context.Background(), common.HexToAddress(addressBob), nil)
	fmt.Printf("User Bob- starting balance: %s\n", balanceBob)




	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
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
		totalSizeFilled += match.SizeFilled
	}

	log.Printf("filled MARKET order  => %d | size:[%.2f]", order.ID, totalSizeFilled)


	return matches, matchesOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	log.Printf("new LIMIT order => type: [%t] || price[%.2f] || size [%.2f]", order.Bid, order.Limit.Price, order.Size)

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
