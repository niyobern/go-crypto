package order

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/adshao/go-binance/v2"
)

func Binance(marketType string, quantity float64, instId string) error {
	client := binance.NewClient(binanceAPIKey, binanceAPISecret)

	symbol := strings.Replace(instId, "-", "", -1)
	side := binance.SideTypeBuy
	orderType := binance.OrderTypeMarket

	var err error
	if marketType == "SPOT" {
		_, err = client.NewCreateOrderService().Symbol(symbol).Side(side).Type(orderType).Quantity(fmt.Sprintf("%f", quantity)).Do(context.Background())
	} else if marketType == "MARGIN" {
		_, err = client.NewCreateMarginOrderService().Symbol(symbol).Side(binance.SideTypeSell).Type(orderType).Quantity(fmt.Sprintf("%f", quantity)).Do(context.Background())
	} else {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func BinanceReverse(marketType string, quantity float64, instId string) error {
	client := binance.NewClient(binanceAPIKey, binanceAPISecret)

	symbol := strings.Replace(instId, "-", "", -1)
	side := binance.SideTypeSell
	orderType := binance.OrderTypeMarket

	var err error
	if marketType == "SPOT" {
		_, err = client.NewCreateOrderService().Symbol(symbol).Side(side).Type(orderType).Quantity(fmt.Sprintf("%f", quantity)).Do(context.Background())
	} else if marketType == "MARGIN" {
		_, err = client.NewCreateMarginOrderService().Symbol(symbol).Side(binance.SideTypeBuy).Type(orderType).Quantity(fmt.Sprintf("%f", quantity)).Do(context.Background())
	} else {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func BinanceLimit(marketType string, price float64, quantity float64, instId string) {
	client := binance.NewClient(binanceAPIKey, binanceAPISecret)

	symbol := strings.Replace(instId, "-", "", -1)
	side := binance.SideTypeBuy
	orderType := binance.OrderTypeLimit

	var order *binance.CreateOrderResponse
	var err error
	if marketType == "SPOT" {
		order, err = client.NewCreateOrderService().Symbol(symbol).Side(side).Type(orderType).Price(fmt.Sprintf("%f", price)).Quantity(fmt.Sprintf("%f", quantity)).TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
	} else if marketType == "MARGIN" {
		order, err = client.NewCreateMarginOrderService().Symbol(symbol).Side(binance.SideTypeBuy).Price(fmt.Sprintf("%f", price)).Type(orderType).Quantity(fmt.Sprintf("%f", quantity)).TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
	} else {
		return
	}
	if err != nil {
		log.Println("Binance order error:", err)
		return
	}

	fmt.Println("Order placed successfully:", order)
}
