package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func assert(t *testing.T, a,b any) {
// 	if !reflect.DeepEqual(a,b) {
// 		t.Errorf("%+v != %+v", a, b)
// 	}
// }

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 1)
	buyOrderB := NewOrder(true, 2)
	buyOrderC := NewOrder(true, 3)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

	fmt.Println(l)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 100)
	sellOrderb := NewOrder(false, 100)
	buyOrder := NewOrder(true, 2000)
	ob.PlaceLimitOrder(10_000, sellOrder)
	ob.PlaceLimitOrder(9_000, buyOrder)
	ob.PlaceLimitOrder(9_000, sellOrderb)

	assert.Equal(t, len(ob.asks), 2)
	assert.Equal(t, len(ob.bids), 1)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 20)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10)
	matches := ob.PlaceMarketOrder(buyOrder)

	fmt.Printf("%+v", matches)
}
