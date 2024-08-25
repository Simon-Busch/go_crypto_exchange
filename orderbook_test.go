package main

import (
	"fmt"
	"testing"
)

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

func TestOrderBook(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(true, 100)
	buyOrderB := NewOrder(true, 2000)
	ob.PlaceOrder(10_000, buyOrderA)
	ob.PlaceOrder(12_000, buyOrderB)

	fmt.Printf("%+v \n", ob)

	for i := 0; i < len(ob.Bids); i++ {
		fmt.Printf("==> %+v\n",ob.Bids[i])
	}
}
