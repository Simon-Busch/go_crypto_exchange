package orderbook

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 1, 0)
	buyOrderB := NewOrder(true, 2, 0)
	buyOrderC := NewOrder(true, 3, 0)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

	assert.Equal(t, len(l.Orders), 2)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 100, 0)
	sellOrderb := NewOrder(false, 100, 0)
	buyOrder := NewOrder(true, 2000, 0)
	ob.PlaceLimitOrder(10_000, sellOrder)
	ob.PlaceLimitOrder(9_000, buyOrder)
	ob.PlaceLimitOrder(9_000, sellOrderb)

	assert.Equal(t, len(ob.Orders), 3)
	assert.Equal(t, ob.Orders[sellOrder.ID], sellOrder)
	assert.Equal(t, ob.Orders[sellOrderb.ID], sellOrderb)

	assert.Equal(t, len(ob.asks), 2)
	assert.Equal(t, len(ob.bids), 1)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 20, 0)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10, 0)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert.Equal(t, len(matches), 1)
	assert.Equal(t, len(ob.asks), 1)
	assert.Equal(t, ob.AskTotalVolume(), 10.0)
	assert.Equal(t, matches[0].Ask, sellOrder)
	assert.Equal(t, matches[0].Bid, buyOrder)
	assert.Equal(t, matches[0].SizeFilled, 10.0)
	assert.Equal(t, matches[0].Price, 10_000.0)
	assert.Equal(t, buyOrder.IsFilled(), true)
}

func TestPlaceMarketOrderMultiFilled(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(true, 5, 0)
	buyOrderB := NewOrder(true, 8, 0)
	buyOrderC := NewOrder(true, 10, 0)
	buyOrderD := NewOrder(true, 1, 0)

	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(10_000, buyOrderA)

	assert.Equal(t, ob.BidTotalVolume(), 24.0)

	sellOrder := NewOrder(false, 20, 0)
	matches := ob.PlaceMarketOrder(sellOrder)

	assert.Equal(t, ob.BidTotalVolume(), 4.0)
	assert.Equal(t, len(matches), 3)
	assert.Equal(t, len(ob.bids), 1)

	fmt.Printf("Matches: %v\n", matches)
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()

	buyOrder := NewOrder(true, 20, 0)
	ob.PlaceLimitOrder(10_000, buyOrder)

	assert.Equal(t, ob.BidTotalVolume(), 20.0)

	ob.CancelOrder(buyOrder)

	assert.Equal(t, ob.BidTotalVolume(), 0.0)

	_, ok := ob.Orders[buyOrder.ID]
	assert.Equal(t, ok, false)
}
