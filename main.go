package main

import (
	"arbitrage/order"
	"arbitrage/transfer"
	"arbitrage/utils"
	"context"
	"log"
	"math"
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
	go func() {
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
	if base != orders.Coin {
		return
	}

	// Channels for tracking success of reverse operations
	buyReverseSuccess := make(chan bool)
	sellReverseSuccess := make(chan bool)

	for _, priceInfo := range priceInfos {
		if priceInfo.Market == "BINANCE" {
			binance = priceInfo
		}
		if priceInfo.Market == "KUCOIN" {
			kucoin = priceInfo
		}
		if binance.Price == 0 || kucoin.Price == 0 {
			return
		}

		if orders.BuyMarket == "BINANCE" {
			if binance.Price >= kucoin.Price {
				var wg sync.WaitGroup
				wg.Add(2)

				go func() {
					defer wg.Done()
					err := order.BinanceReverse("SPOT", orders.Amount, instId)
					if err != nil {
						log.Fatalf("Binance failed to reverse sell: %f %s", orders.Amount, instId)
						buyReverseSuccess <- false
					} else {
						buyReverseSuccess <- true
					}
				}()

				go func() {
					defer wg.Done()
					err := order.KucoinReverse("MARGIN", instId, orders.Amount)
					if err != nil {
						log.Fatalf("Kucoin failed to reverse buy: %f %s", orders.Amount, instId)
						sellReverseSuccess <- false
					} else {
						sellReverseSuccess <- true
					}
				}()

				wg.Wait()

				if <-buyReverseSuccess && <-sellReverseSuccess {
					time.Sleep(1 * time.Second)
					transfer.KucoinRepayLoan(base, orders.Amount)
					time.Sleep(1 * time.Second)
					_, err := transfer.KucoinMargin2spot("USDT", CAPITAL)
					if err != nil {
						log.Fatalln("Kucoin transfer error:", err)
					}
					err = utils.SaveOrders(db, "", "", "", 0, 0, 0)
					if err != nil {
						log.Fatalln("Save error:", err)
					} else {
						log.Fatalln("Save error:", err)
					}
				}
			}
		} else {
			if kucoin.Price >= binance.Price {
				var wg sync.WaitGroup
				wg.Add(2)

				go func() {
					defer wg.Done()
					err := order.KucoinReverse("SPOT", instId, orders.Amount)
					if err != nil {
						log.Fatalf("Kucoin failed to reverse sell: %f %s", orders.Amount, instId)
						buyReverseSuccess <- false
					} else {
						buyReverseSuccess <- true
					}
				}()

				go func() {
					defer wg.Done()
					err := order.BinanceReverse("MARGIN", orders.Amount, instId)
					if err != nil {
						log.Fatalf("Binance failed to reverse buy: %f %s", orders.Amount, instId)
						sellReverseSuccess <- false
					} else {
						sellReverseSuccess <- true
					}
				}()

				wg.Wait()

				if <-buyReverseSuccess && <-sellReverseSuccess {
					time.Sleep(1 * time.Second)
					transfer.BinanceRepayMarginLoan(base, orders.Amount)
					time.Sleep(1 * time.Second)
					_, err := transfer.BinanceMargin2Spot("USDT", CAPITAL)
					if err != nil {
						log.Fatalln("Transfer error:", err)
					}
					err = utils.SaveOrders(db, "", "", "", 0, 0, 0)
					if err != nil {
						log.Fatalln("Save error:", err)
					}else {
						log.Fatalln("Save error:", err)
					}
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
		if final > CAPITAL+0.1 && minPrice.Market == "BINANCE" && maxPrice.Market == "KUCOIN" {
			makeOrders(isOpen, db, orders, instId, "BINANCE", minPrice.Price, maxPrice.Price)
		}
		if final > CAPITAL+0.1 && minPrice.Market == "KUCOIN" && maxPrice.Market == "BINANCE" {
			makeOrders(isOpen, db, orders, instId, "KUCOIN", minPrice.Price, maxPrice.Price)
		}
	}
}

func makeOrders(isOpen *bool, db *leveldb.DB, orders *utils.OrderData, instId string, buyMarket string minPrice, maxPrice float64) {
	if *isOpen {
		time.Sleep(100 * time.Millisecond)
		return
	}
	buyAmount := math.Floor(CAPITAL / minPrice)
	if buyAmount < CAPITAL {
		return
	}
	base := strings.Split(instId, "-")[0]

	var wg sync.WaitGroup
	buySuccess := make(chan bool)
	sellSuccess := make(chan bool)

	if buyMarket == "BINANCE" {
		wg.Add(2)

		go func() {
			defer wg.Done()
			err := order.Binance("SPOT", buyAmount, instId)
			if err != nil {
				log.Fatalf("Failed to place Binance buy order: %v", err)
				buySuccess <- false
			} else {
				buySuccess <- true
			}
		}()

		go func() {
			defer wg.Done()
			_, err := transfer.KucoinSpot2Margin("USDT", CAPITAL)
			if err != nil {
				log.Fatalln("Transfer error:", err)
			}
			err = order.Kucoin("MARGIN", instId, buyAmount)
			if err != nil {
				log.Fatalf("Failed to place Kucoin short order: %v", err)
				sellSuccess <- false
			} else {
				sellSuccess <- true
			}
		}()

		wg.Wait()

		if <-buySuccess && <-sellSuccess {
			time.Sleep(1 * time.Second)
			*isOpen = true
			orders.BuyMarket = "BINANCE"
			orders.SellMarket = "KUCOIN"
			orders.Amount = buyAmount
			orders.Coin = base
			err := utils.SaveOrders(db, "BINANCE", "KUCOIN", base, buyAmount, minPrice, maxPrice)
			if err != nil {
				log.Fatalln("Save error:", err)
			}
		}
	} else {
		wg.Add(2)

		go func() {
			defer wg.Done()
			err := order.Kucoin("SPOT", instId, buyAmount)
			if err != nil {
				log.Fatalf("Failed to place Kucoin buy order: %v", err)
				buySuccess <- false
			} else {
				buySuccess <- true
			}
		}()

		go func() {
			defer wg.Done()
			_, err := transfer.BinanceSpot2Margin("USDT", CAPITAL)
			if err != nil {
				log.Fatalln("Transfer error:", err)
			}
			err = order.Binance("MARGIN", buyAmount, "MARGIN")
			if err != nil {
				log.Fatalf("Failed to place Binance short order: %v", err)
				sellSuccess <- false
			} else {
				sellSuccess <- true
			}
		}()

		wg.Wait()

		if <-buySuccess && <-sellSuccess {
			time.Sleep(1 * time.Second)
			*isOpen = true
			orders.BuyMarket = "KUCOIN"
			orders.SellMarket = "BINANCE"
			orders.Amount = buyAmount
			orders.Coin = base
			err := utils.SaveOrders(db, "KUCOIN", "BINANCE", base, buyAmount, minPrice, maxPrice)
			if err != nil {
				log.Fatalln("Save error:", err)
			}
		}
	}
}
