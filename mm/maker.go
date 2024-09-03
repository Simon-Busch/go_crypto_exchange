package mm

import (
	"time"

	"github.com/Simon-Busch/go_crypto_exchange/client"
	"github.com/sirupsen/logrus"
)

type Config struct {
	UserID 					int64
	OrderSize 			float64
	MinSpread 			float64
	SeedOffset 			float64
	ExchangeClient 	*client.Client
	MakeInterval 		time.Duration
}

type MarketMaker struct {
	userID 					int64
	orderSize 			float64
	minSpread 			float64
	seedOffset 			float64
	exchangeClient 	*client.Client
	makeInterval 		time.Duration
}


func NewMarketMaker(cfc Config) *MarketMaker {
	return &MarketMaker{
		userID: 				cfc.UserID,
		orderSize: 			cfc.OrderSize,
		minSpread: 			cfc.MinSpread,
		seedOffset: 		cfc.SeedOffset,
		exchangeClient: cfc.ExchangeClient,
		makeInterval: 	cfc.MakeInterval,
	}
}

func (mm *MarketMaker) Start() {
	logrus.WithFields(logrus.Fields{
		"userID": 			mm.userID,
		"orderSize": 		mm.orderSize,
		"minSpread": 		mm.minSpread,
		"makeInterval": mm.makeInterval,
	}).Info("Starting market maker")
	go mm.makerLoop()
}

func (mm *MarketMaker) makerLoop() {
	ticker := time.NewTicker(mm.makeInterval)
	for {

		bestBid, err := mm.exchangeClient.GetBestBid()
		if err != nil {
			logrus.Error(err)
			break;
		}

		bestAsk, err := mm.exchangeClient.GetBestAsk()
		if err != nil {
			logrus.Error(err)
			break;
		}

		if bestAsk == 0.0 && bestBid == 0.0 {
			if err := mm.seedMarket(); err != nil {
				logrus.Error(err)
				break;
			}
		}

		<- ticker.C
	}
}

func (mm *MarketMaker) seedMarket() error {
	currentPrice := simulateFetchCurrentETHPrice()

	logrus.WithFields(logrus.Fields{
		"currentETHPrice": currentPrice,
		"seedOffset": 			mm.seedOffset,
	}).Info("Orderbook empty --> seeding market")

	bidOrder := &client.PlaceOrderParams{
		UserID: 			mm.userID,
		Bid: 					true,
		Size: 				mm.orderSize,
		Price:				currentPrice - mm.seedOffset,
	}

	_, err := mm.exchangeClient.PlaceLimitOrder(bidOrder)
	if err != nil {
		return err
	}

	askOrder := &client.PlaceOrderParams{
		UserID: 			mm.userID,
		Bid: 					false,
		Size: 				mm.orderSize,
		Price:				currentPrice + mm.seedOffset,
	}

	_, err = mm.exchangeClient.PlaceLimitOrder(askOrder)
	if err != nil {
		return err
	}

	return nil
}

// This will simulate a call to an order exchange to fetch
// the current ETH price so we can offset both for bids and asks
func simulateFetchCurrentETHPrice() float64 {
	time.Sleep(80 * time.Millisecond)

	return 1000.0
}
