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
	return fmt.Sprintf("Order{Size: %f, Bid: %t}", o.Size, o.Bid)
}

func (o *Order) IsFilled() bool {
	return o.Size == 0.0
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

func (l *Limit) Fill(o *Order) []Match {
	matches := []Match{}

	for _, order := range l.Orders {
		match := l.fillOrder(order, o)
		matches = append(matches, match)

		if o.IsFilled() {
			break
		}
	}

	return matches
}

func (l *Limit) fillOrder(a, b *Order) Match {
	var (
		bid 				*Order
		ask 				*Order
		sizeFilled 	float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if a.Size > b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0.0
	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0.0
	}

	return Match{
		Ask: 				ask,
		Bid: 				bid,
		SizeFiled: 	sizeFilled,
		Price: 			l.Price,
	}
}

type Orderbook struct {
	asks 					[]*Limit
	bids 					[]*Limit

	AskLimits 		map[float64]*Limit
	BidsLimits 		map[float64]*Limit
}

func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	matches := []Match{}

	if o.Bid {
		if o.Size > ob.AskTotalVolumes() {
			panic(fmt.Errorf("not enough volume [size: %.2f] for order [size :%.2f]", ob.AskTotalVolumes(), o.Size))
		}

		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)
		}
	} else {
		if o.Size > ob.BidTotalVolumes() {
			panic("not enough volume sitting in the books")
		}

		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)
		}
	}

	return matches
}

func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit
	if o.Bid {
		limit = ob.BidsLimits[price]
		if limit == nil {
			limit = NewLimit(price)
			ob.bids = append(ob.bids, limit)
			ob.BidsLimits[price] = limit
		}
	} else {
		limit = ob.AskLimits[price]
		if limit == nil {
			limit = NewLimit(price)
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	limit.AddOrder(o)
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks: 			[]*Limit{},
		bids: 			[]*Limit{},
		AskLimits: 	make(map[float64]*Limit),
		BidsLimits: make(map[float64]*Limit),
	}
}

func (ob *Orderbook) BidTotalVolumes() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].TotalVolume
	}

	return totalVolume
}

func (ob *Orderbook) AskTotalVolumes() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].TotalVolume
	}

	return totalVolume
}

func (ob *Orderbook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

func (ob *Orderbook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}
