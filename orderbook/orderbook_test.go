package orderbook

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

	fmt.Println(l)
}

// Uncomment to test PlaceOrder() if place limit/market order functions are not yet implemented
//
// func TestOrderbook(t *testing.T) {
// 	ob := NewOrderbook()
// 	buyOrderA := NewOrder(true, 10)
// 	buyOrderB := NewOrder(true, 2000)
// 	ob.PlaceOrder(18_000, buyOrderA)
// 	ob.PlaceOrder(19_000, buyOrderB)
// 	fmt.Printf("%+v", ob)
// 	// for i := 0; i < len(ob.Bids); i++ {
// 	// 	fmt.Printf("%+v", ob.Bids[i])
// 	// }
// 	// fmt.Println(ob.Bids)
// }

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := NewOrder(false, 10)
	sellOrderB := NewOrder(false, 5)
	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(18_000, sellOrderB)

	assert(t, len(ob.asks), 2)

}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 20)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)

	fmt.Printf("%+v", matches)

}

func TestPlaceMarketOrderMultiFilled(t *testing.T) {

	// Create orderbook
	// Create and place 23 buy orders, 10 at 10k, 8 at 9k, 5 at 5k
	// Check bids total volume (assert())

	ob := NewOrderbook()

	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)
	buyOrderD := NewOrder(true, 1)

	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(10_000, buyOrderD)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(5_000, buyOrderC)

	assert(t, ob.BidTotalVolume(), 24.00)

	// Create and place 20 sell orders
	// match buy orders and sell order
	// Check asks total remaining volume (assert())

	sellOrder := NewOrder(false, 20)
	matches := ob.PlaceMarketOrder(sellOrder)

	assert(t, ob.BidTotalVolume(), 4.0)
	assert(t, len(matches), 3)
	assert(t, len(ob.bids), 1)

	fmt.Printf("%+v", matches)
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()
	buyOrder := NewOrder(true, 4.0)
	ob.PlaceLimitOrder(10_000, buyOrder)

	assert(t, ob.BidTotalVolume(), 4.0)
	// Test cancel order
	ob.CancelOrder(buyOrder)
	assert(t, ob.BidTotalVolume(), 0.0)
}
