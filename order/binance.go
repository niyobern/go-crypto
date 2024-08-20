package order

import (
	"context"
	"strings"
	"fmt"
	"arbitrage/utils"
	"github.com/adshao/go-binance/v2"
)

func Binance(orders chan Order, marketType string, quantity float64, instId string) {
	apiKey := "JnVmPqBo5bSZX6NQP5EMJTukpmgcPAJdrirAFGWTfCuJfIeHqHbcrbMIfdiOWpVR"
	secretKey := "XKFk5XfmKrTdw4nA8ubFiufyij6uMK1EHDzmNcxALn9dSscaz7kdh6aa0fygUqFl"
	
	trie := utils.Initialize()

	client := binance.NewClient(apiKey, secretKey)

	symbol := strings.Replace(instId, "-", "", -1)
	side := binance.SideTypeBuy
	orderType := binance.OrderTypeMarket
	
    var order *binance.CreateOrderResponse
	var err error
    if marketType == "SPOT" {
		order, err = client.NewCreateOrderService().Symbol(symbol).Side(side).Type(orderType).QuoteOrderQty(fmt.Sprintf("%f", quantity)).Do(context.Background())
	} else if marketType == "MARGIN" {
		order, err = client.NewCreateMarginOrderService().Symbol(symbol).Side(binance.SideTypeSell).Type(orderType).QuoteOrderQty(fmt.Sprintf("%f", quantity)).Do(context.Background())
	} else {
		return
	}
	if err != nil {
		fmt.Println("Failed to place order:", err)
		return
	}

	orders <- Order{
		Symbol: utils.GetQuote(order.Symbol, trie),
		OrderID: order.OrderID,
		Price: order.Price,
		Quantity: order.ExecutedQuantity,
		Status: string(order.Status),
		Side: string(order.Side),
		Time: order.TransactTime,
	}

	fmt.Println("Order placed successfully:", order)
	
}