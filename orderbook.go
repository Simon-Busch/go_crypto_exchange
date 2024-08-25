package main

import (
	"fmt"
	"time"
)

type Match struct {
	Ask 			*Order
	Bid 			*Order
	SizeFiled float64
	Price 		float64
}

type Order struct {
	Size 					float64
	Bid 					bool
	Limit 				*Limit
	Timestamp 		int64
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:				size,
		Bid: 				bid,
		Limit: 			nil,
		Timestamp: 	time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("Order{Size: %f, Bid: %t, Timestamp: %d}", o.Size, o.Bid, o.Timestamp)
}

// Group of order at a certain price level
type Limit struct {
	Price  				float64
	Orders 				[]*Order
	TotalVolume 	float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price: 			price,
		Orders:			[]*Order{}, // using a slice because we don't know yet how many order there will be
	}
}

func (l *Limit) String() string {
	return fmt.Sprintf("Limit{Price: %.2f | TotalVolume: %.2f | Orders: %v}", l.Price, l.TotalVolume, l.Orders)
}

func (l *Limit) AddOrder(order *Order) {
	order.Limit = l
	l.Orders = append(l.Orders, order)
	l.TotalVolume += order.Size
}

func (l *Limit) DeleteOrder(order *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == order {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			break
		}
	}

	order.Limit = nil
	l.TotalVolume -= order.Size
}

type Orderbook struct {
	Asks 					[]*Limit
	Bids 					[]*Limit

	AskLimits 		map[float64]*Limit
	BidsLimits 		map[float64]*Limit
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		Asks: 			[]*Limit{},
		Bids: 			[]*Limit{},
		AskLimits: 	make(map[float64]*Limit),
		BidsLimits: make(map[float64]*Limit),
	}
}

func (ob *Orderbook) PlaceOrder(price float64, o *Order) []Match {
	// 1. Try to match the order
	// Matching logic

	// 2. add the rest of the order to the books
	if o.Size > 0.0 {
		ob.add(price, o)
	}

	return []Match{}
}

func (ob *Orderbook) add(price float64, o *Order) {
	var limit *Limit
	if o.Bid {
		limit = ob.BidsLimits[price]
		if limit == nil {
			limit = NewLimit(price)
			ob.Bids = append(ob.Bids, limit)
			ob.BidsLimits[price] = limit
		}
	} else {
		limit = ob.AskLimits[price]
		if limit == nil {
			limit = NewLimit(price)
			ob.Asks = append(ob.Asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	limit.AddOrder(o)
}
