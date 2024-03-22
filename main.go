package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/larrysammii/cryto-exchange-go/orderbook"
)

const (
	// Not a real private key of course, it's from ganache
	exchangePrivateKey string    = "a8e2d0c7ce9b08546f647f5a2a0f4927feea43dd070b30e1a08e7d3cd9ab48bb"
	MarketOrder        OrderType = "MARKET"
	LimitOrder         OrderType = "LIMIT"
	MarketETH          Market    = "ETH"
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler
	ex := NewExchange()

	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)

	// ------------------------------------------------------------------------------
	// use ganache for the following

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	// address from ganache
	// address := common.HexToAddress("0xCCDebc03CdF5EE3d56bDA59C72407E339084333F")
	// balance, err := client.BalanceAt(ctx, address, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(balance)

	// private key also from ganache
	privateKey, err := crypto.HexToECDSA("a8e2d0c7ce9b08546f647f5a2a0f4927feea43dd070b30e1a08e7d3cd9ab48bb")
	if err != nil {
		log.Fatal(err)
	}
	// ------------------------------------------------------------------------------

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(100000000000000000) // in wei (1 eth)
	gasLimit := uint64(21000)               // in units

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0xa1dB4A8676C9577B02FD1Fed74DE88EDc6BA86d3")
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)
	// chainID, err := client.NetworkID(ctx) // 1337
	// if err != nil {
	// 	log.Fatal(err)
	// }
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	if err = client.SendTransaction(ctx, signedTx); err != nil {
		log.Fatal(err)
	}

	balance, err := client.BalanceAt(ctx, toAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(balance)

	e.Start(":3000")

}

type User struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(privateKey string) *User {

	privateKeyNewUser, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}
	return &User{
		PrivateKey: privateKeyNewUser,
	}
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

type OrderType string

type Market string

type Exchange struct {
	PrivateKey *ecdsa.PrivateKey
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange() *Exchange {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	privateKey, err := crypto.HexToECDSA(exchangePrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	return &Exchange{
		PrivateKey: privateKey,
		orderbooks: orderbooks,
	}
}

type PlaceOrderRequest struct {
	Type   OrderType // limit or market
	Bid    bool
	Size   float64
	Price  float64
	Market Market
}

type Order struct {
	ID        int64
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type OrderbookData struct {
	TotalBidVolume float64
	TotalAskVolume float64
	Asks           []*Order
	Bids           []*Order
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"message": "market not found"})
	}

	orderbookData := OrderbookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}
	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)

		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)

		}
	}
	return c.JSON(http.StatusOK, orderbookData)
}

func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketETH]
	order := ob.Orders[int64(id)]
	ob.CancelOrder(order)

	return c.JSON(200, map[string]any{"message": "order canceled"})
}

type MatchedOrder struct {
	Price float64
	Size  float64
	ID    int64
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)
	ob := ex.orderbooks[market]
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size)

	if placeOrderData.Type == LimitOrder {
		ob.PlaceLimitOrder(placeOrderData.Price, order)
		return c.JSON(200, map[string]any{"message": "limit order placed"})
	}

	if placeOrderData.Type == MarketOrder {
		matches := ob.PlaceMarketOrder(order)
		matchedOrders := make([]*MatchedOrder, len(matches))

		isBid := false
		if order.Bid {
			isBid = true
		}

		for i := 0; i < len(matches); i++ {
			id := matches[i].Bid.ID

			if isBid {
				id = matches[i].Ask.ID
			}

			matchedOrders[i] = &MatchedOrder{
				ID:    id,
				Size:  matches[i].SizeFilled,
				Price: matches[i].Price,
			}

		}
		return c.JSON(200, map[string]any{"matches": matchedOrders})
	}

	return nil
}
