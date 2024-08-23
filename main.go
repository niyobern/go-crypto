package main

import (
	"arbitrage/order"
	"arbitrage/transfer"
	"arbitrage/utils"
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
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

const CAPITAL = 60.0

func main() {
	tickers := make(chan TickerGeneral)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
    
	db, err := utils.Database("orderdb")
    if err != nil {
        log.Fatal("Failed to open LevelDB:", err)
    }
    defer db.Close()

	orders, isOpen, err := utils.FindOpenOrders(db)

	if err != nil {
		log.Fatal("Failed to get orders:", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		go Kucoin(ctx, tickers)
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
				go CheckifOrderOpen(db, orders, &isOpen, ticker.InstId, prices[ticker.InstId])
				checkArbitrage(&isOpen, db, &orders, ticker.InstId, prices[ticker.InstId])
			}
			mu.Unlock()
		}
	}()

	// Wait for the WebSocket goroutines to finish
	wg.Wait()
	log.Println("shutting down")
}

func CheckifOrderOpen(db *leveldb.DB, orders utils.OrderData, isOpen *bool, instId string, priceInfos map[string]PriceInfo) {
	var binance, kucoin PriceInfo
	base := strings.Split(instId, "-")[0]
	for _, priceInfo := range priceInfos {
		
		if priceInfo.Market == "BINANCE" {
			binance = priceInfo
		}
		if priceInfo.Market == "KUCOIN" {
			kucoin = priceInfo
		}

		if orders.BuyMarket  == "BINANCE" {
			if binance.Price >= kucoin.Price {
				go order.BinanceReverse("SPOT", orders.Amount, instId)
				go order.KucoinReverse("MARGIN", instId, orders.Amount)
				time.Sleep(1 * time.Second)
				transfer.KucoinRepayLoan("USDT", CAPITAL)
				time.Sleep(1 * time.Second)
				transfer.KucoinMargin2spot("USDT", CAPITAL)
				err := utils.SaveBuyrders(db, "", "", 0, 0)
				if err != nil {
					log.Fatal(err)
				}
				err = utils.SaveSellOrders(db, "", 0)
				if err == nil {
					*isOpen = false
				} else {
					log.Fatal(err)
				}
				
			}
		} else {
			if kucoin.Price >= binance.Price {
                go order.KucoinReverse("SPOT", instId, orders.Amount)
				go order.BinanceReverse("MARGIN", orders.Amount, instId)
				time.Sleep(1 * time.Second)
				transfer.BinanceRepayMarginLoan(base, CAPITAL)
				time.Sleep(1 * time.Second)
				transfer.BinanceMargin2Spot("USDT", CAPITAL)
				err := utils.SaveBuyrders(db, "", "", 0, 0)
				if err != nil {
					log.Fatal(err)
				}
				err = utils.SaveSellOrders(db, "", 0)
				if err == nil {
					*isOpen = false
				} else {
					log.Fatal(err)
				}
			}
		}
	}
}

func checkArbitrage(isOpen *bool, db *leveldb.DB, orders *utils.OrderData, instId string, priceInfos map[string]PriceInfo) {
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
        fees := 0.001 // Assume 0.1% fees for each trade
		coeficient := CAPITAL / (minPrice.Price + (fees * minPrice.Price))
		sellValue := coeficient * (maxPrice.Price - (fees * maxPrice.Price))
		final := sellValue - 0.08 // Considered transfer fees to be $2
		if final > CAPITAL + 0.1 && minPrice.Market == "BINANCE" && maxPrice.Market == "KUCOIN" {
			// log.Printf("Arbitrage for %s: Buy on %s at %.4f, sell on %s at %.4f for %.2f USDT, get %.2f\n",
			// instId, minPrice.Market, minPrice.Price, maxPrice.Market, maxPrice.Price, CAPITAL, final)
			makeOrders(isOpen, db, orders, instId, "BINANCE", "KUCOIN", minPrice.Price, maxPrice.Price)
		}
		if final > CAPITAL + 0.1 && minPrice.Market == "KUCOIN" && maxPrice.Market == "BINANCE"  {
			// log.Printf("Arbitrage for %s: Buy on %s at %.4f, sell on %s at %.4f for %.2f USDT, get %.2f\n",
			// instId, minPrice.Market, minPrice.Price, maxPrice.Market, maxPrice.Price, CAPITAL, final)
			makeOrders(isOpen, db, orders, instId, "KUCOIN", "BINANCE", minPrice.Price, maxPrice.Price)
		}
	}
}

func makeOrders(isOpen *bool, db *leveldb.DB, orders *utils.OrderData, instId string, buyMarket string, sellMarket string, minPrice, maxPrice float64){
	if *isOpen {
		return
	}
	buyAmount := CAPITAL/maxPrice
	base := strings.Split(instId, "-")[0]
	switch buyMarket {
		case "BINANCE":
			go func(){
				order.Binance("SPOT", buyAmount, instId)
				utils.SaveBuyrders(db, "BINANCE", base, minPrice, buyAmount)
				orders.Amount = buyAmount
				orders.BuyMarket = "BINANCE"
				orders.Coin = base
				*isOpen = true
			}()

		case "KUCOIN":
			go func(){
				order.Kucoin("SPOT", instId, buyAmount)
				utils.SaveBuyrders(db, "KUCOIN", base, minPrice, buyAmount)
				orders.Amount = buyAmount
				orders.BuyMarket = "KUCOIN"
				orders.Coin = base
				*isOpen = true
			}()
	}
	switch sellMarket {
		case "BINANCE":
			go func(){
				_, err := transfer.BinanceSpot2Margin("USDT", CAPITAL)
				if err != nil {
					transfer.BinanceMargin2Spot("USDT", CAPITAL)
					log.Println("BiS2M error", err)
					return
				}
				order.Binance("MARGIN", buyAmount, instId)
				utils.SaveSellOrders(db, "BINANCE", maxPrice)
				orders.SellMarket = "BINANCE"
				orders.SellPrice = maxPrice
			}()
		case "KUCOIN":
			go func(){
				_, err := transfer.KucoinSpot2Margin("USDT", CAPITAL)
				if err != nil {
					transfer.KucoinMargin2spot("USDT", CAPITAL)
					log.Println("KuS2M error", err)
					return
				}
				order.Kucoin("MARGIN", instId, buyAmount)
				utils.SaveSellOrders(db, "KUCOIN", maxPrice)
				orders.SellMarket = "KUCOIN"
				orders.SellPrice = maxPrice
			}()
	}
}