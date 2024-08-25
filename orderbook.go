package main

import (
	"fmt"
	"sort"
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

type Orders []*Order

func (o Orders) Len() int 						{ return len(o) }
func (o Orders) Swap(i,j int) 				{ o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i,j int) bool 		{ return o[i].Timestamp < o[j].Timestamp }

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
	Orders 				Orders
	TotalVolume 	float64
}

type Limits []*Limit

type ByBestAsk struct { Limits }

func (a ByBestAsk) Len() int 						{ return len(a.Limits) }
func (a ByBestAsk) Swap(i,j int) 				{ a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i,j int) bool 	{ return a.Limits[i].Price < a.Limits[j].Price }

type ByBestBid struct { Limits }

func (a ByBestBid) Len() int 						{ return len(a.Limits) }
func (a ByBestBid) Swap(i,j int) 				{ a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestBid) Less(i,j int) bool 	{ return a.Limits[i].Price > a.Limits[j].Price }

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

	sort.Sort(l.Orders)
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
