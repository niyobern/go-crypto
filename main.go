package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
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

func order() {
	tickers := make(chan TickerGeneral)
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		okx(ctx, tickers)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kucoin(ctx, tickers)
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
				checkArbitrage(ticker.InstId, prices[ticker.InstId])
			}
            checkTriangularArbitrage(prices)
			mu.Unlock()
		}
	}()

	// Wait for the WebSocket goroutines to finish
	wg.Wait()
	log.Println("shutting down")
}

func checkArbitrage(instId string, priceInfos map[string]PriceInfo) {
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
        fees := 0.1 // Assume 0.1% fees for each trade
        gain := (maxPrice.Price - minPrice.Price ) * 100 / minPrice.Price
		profit := gain - (fees * 2)
		coeficient := 2000 / (minPrice.Price + (0.001 * minPrice.Price))
		sellValue := coeficient * (maxPrice.Price - (0.001 * maxPrice.Price))
		if minPrice.Market == "BINANCE" {
            log.Printf("Arbitrage for %s: Buy on %s at %.4f, sell on %s at %.4f for profit %.2f%%, for 2000 USDT, get %.2f\n",
			instId, minPrice.Market, minPrice.Price, maxPrice.Market, maxPrice.Price, profit, sellValue)
        }
	}
}


func checkTriangularArbitrage(prices map[string]map[string]PriceInfo) {
	for instId1, markets1 := range prices {
		for instId2, markets2 := range prices {
			if instId1 == instId2 {
				continue
			}

			base1, quote1 := parseInstId(instId1)
			base2, quote2 := parseInstId(instId2)

			if base1 == quote2 {
				for instId3, markets3 := range prices {
					if instId3 == instId1 || instId3 == instId2 {
						continue
					}

					base3, quote3 := parseInstId(instId3)

					if base3 == quote1 && quote3 == base2 {
						// We have base1 -> quote1, base1 -> base2 (base2 == quote2), quote1 -> base2 (quote1 == base3)
						for _, price1 := range markets1 {
							for _, price2 := range markets2 {
								for _, price3 := range markets3 {
									profit := 1 / price1.Price * price2.Price * price3.Price
									if profit > 1 {
										log.Printf("Triangular arbitrage opportunity: %s -> %s -> %s -> %s with profit %.2f%%\n",
											instId1, instId2, instId3, instId1, (profit-1)*100)
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func parseInstId(instId string) (string, string) {
	parts := strings.Split(instId, "-")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}