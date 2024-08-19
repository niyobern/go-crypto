package main

import (
	"arbitrage/balance"
	"arbitrage/order"
	"arbitrage/transfer"
	"arbitrage/utils"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

type TickerGeneral struct {
	InstId string `json:"instId"`
	Price  string `json:"price"`
	Market string `json:"Market"`
}

type PriceInfo struct {
	Price  float64
	Market string
}

const CAPITAL = 1000.0

func main() {
	tickers := make(chan TickerGeneral)
	orders := make(chan order.Order)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
    
	db, err := utils.Database("orderdb")
    if err != nil {
        log.Fatal("Failed to open LevelDB:", err)
    }
    defer db.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		go Kucoin(ctx, tickers)
	}()

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		go func() {
			for {
				select {
				case <-ticker.C:
					// For Binance, transfer all received coins to margin account and repay the margin loan
					coins, err := balance.Binance()
					if err != nil {
						log.Fatalf("Error fetching Binance account info: %v", err)
					}
					for _, coin := range coins {
						if coin.Currency != "USDT" && coin.Wallet == "FUNDING" {
							// pay the coin to the margin loan
							transfer.BinanceFunding2Spot(coin.Currency, coin.Balance)
							time.Sleep(10 * time.Second)
							transfer.BinanceSpot2Margin(coin.Currency, coin.Balance)
							time.Sleep(10 * time.Second)
							transfer.BinanceRepayMarginLoan(coin.Currency, coin.Balance)
						}
					}
					transfer.BinanceMargin2Spot("USDT", CAPITAL)
					transfer.KucoinMargin2spot("USDT", CAPITAL)
				case <-ctx.Done():
					return
				}
			}
		}()
	}()

	wg.Add(1)
	go func () {
		defer wg.Done()
		binance(ctx, tickers)
	}()

	// Handle interrupt signals for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("interrupt signal received")
		cancel()
	}()

	// Store latest prices from each exchange
	prices := make(map[string]map[string]PriceInfo) // InstId -> Market -> PriceInfo
	var mu sync.Mutex

	// Process tickers
	go func() {
		for ticker := range tickers {
			price, err := strconv.ParseFloat(ticker.Price, 64)
			if err != nil {
				log.Printf("invalid price %s for %s: %v", ticker.Price, ticker.InstId, err)
				continue
			}

			mu.Lock()
			if _, exists := prices[ticker.InstId]; !exists {
				prices[ticker.InstId] = make(map[string]PriceInfo)
			}
			prices[ticker.InstId][ticker.Market] = PriceInfo{Price: price, Market: ticker.Market}

			// Check for arbitrage opportunities
			if len(prices[ticker.InstId]) > 1 {
				checkArbitrage(ticker.InstId, prices[ticker.InstId], orders)
			}
			mu.Unlock()
		}
	}()

	go func (){
		for order := range orders {
			fmt.Println(order)
		}
	}()

	// Wait for the WebSocket goroutines to finish
	wg.Wait()
	log.Println("shutting down")
}

func checkArbitrage(instId string, priceInfos map[string]PriceInfo, orders chan order.Order) {
	var minPrice, maxPrice PriceInfo
	first := true
	for _, priceInfo := range priceInfos {
		if first {
			minPrice = priceInfo
			maxPrice = priceInfo
			first = false
		} else {
			if priceInfo.Price < minPrice.Price {
				minPrice = priceInfo
			}
			if priceInfo.Price > maxPrice.Price {
				maxPrice = priceInfo
			}
		}
	}

	if minPrice.Market != maxPrice.Market {
        fees := 0.0015 // Assume 0.1% fees for each trade
		coeficient := CAPITAL / (minPrice.Price + (fees * minPrice.Price))
		sellValue := coeficient * (maxPrice.Price - (fees * maxPrice.Price))
		final := sellValue - 2
		if final > 1002 && minPrice.Market == "BINANCE" && maxPrice.Market == "KUCOIN"  {
			makeOrders(orders, instId, "BINAooooNCE", "KUoooCOIN", coeficient*1.25)
			log.Printf("Arbitrage for %s: Buy on %s at %.4f, sell on %s at %.4f for 1000 USDT, get %.2f\n",
			instId, minPrice.Market, minPrice.Price, maxPrice.Market, maxPrice.Price, final)
		}
	}
}

func makeOrders(orders chan order.Order, instId string, buyMarket string, sellMarket string, sellAmount float64){
	switch buyMarket {
		case "BINANCE":
			go order.Binance(orders, "SPOT", CAPITAL, instId)
		case "KUCOIN":
			go order.Kucoin("MARGIN", instId, sellAmount)
	}
	switch sellMarket {
		case "BINANCE":
			go order.Binance(orders, "MARGIN", sellAmount, instId)
		case "KUCOIN":
			go order.Kucoin("SPOT", instId, CAPITAL)
	}
}