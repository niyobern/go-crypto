package main

import (
	"arbitrage/utils"
	"context"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// Define structures
type BinanceTicker struct {
	UpdateID int64  `json:"u"`
	Symbol   string `json:"s"`
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
}

func binance(ctx context.Context, tickers chan TickerGeneral) {
	trie := utils.Initialize()

	// Binance WebSocket endpoint for order book data - using specific symbols
	endpoint := "wss://stream.binance.com:9443/ws/" +
		"cfxusdt@bookTicker/celousdt@bookTicker/flowusdt@bookTicker/icpusdt@bookTicker/icxusdt@bookTicker/filusdt@bookTicker/xlmusdt@bookTicker/algousdt@bookTicker/enjusdt@bookTicker/dgbusdt@bookTicker/oneusdt@bookTicker/iostusdt@bookTicker/rvnusdt@bookTicker/zilusdt@bookTicker/waxpusdt@bookTicker/aceusdt@bookTicker/egldusdt@bookTicker/gasusdt@bookTicker/hbarusdt@bookTicker/eosusdt@bookTicker/dydxusdt@bookTicker/movrusdt@bookTicker/polusdt@bookTicker/wldusdt@bookTicker/aptusdt@bookTicker/iotausdt@bookTicker/dogesusdt@bookTicker/xtzusdt@bookTicker/nearusdt@bookTicker/ltcusdt@bookTicker/atomusdt@bookTicker/rdntusdt@bookTicker/arbusdt@bookTicker/magicusdt@bookTicker/neousdt@bookTicker/usdcusdt@bookTicker/tonusdt@bookTicker/xrpusdt@bookTicker/ethusdt@bookTicker/qtumusdt@bookTicker/suiusdt@bookTicker/gmtusdt@bookTicker/catiusdt@bookTicker/opusdt@bookTicker/etcusdt@bookTicker/arusdt@bookTicker/hmstrusdt@bookTicker/flmusdt@bookTicker/avaxusdt@bookTicker/thetausdt@bookTicker/trxusdt@bookTicker/ontusdt@bookTicker/ksmusdt@bookTicker/injusdt@bookTicker/notusdt@bookTicker/gmxusdt@bookTicker/bchusdt@bookTicker/adausdt@bookTicker/dotusdt@bookTicker/zrousdt@bookTicker/elfusdt@bookTicker/metisusdt@bookTicker/minausdt@bookTicker/tiausdt@bookTicker/joeusdt@bookTicker/dogeusdt@bookTicker/strkusdt@bookTicker/renderusdt@bookTicker/bnbusdt@bookTicker/flokiusdt@bookTicker/jstusdt@bookTicker/tnsrusdt@bookTicker/wusdt@bookTicker/jtousdt@bookTicker/bonkusdt@bookTicker/pythusdt@bookTicker/wifusdt@bookTicker/jupusdt@bookTicker/solusdt@bookTicker/rayusdt@bookTicker/bomeusdt@bookTicker"

	// Connect to the WebSocket
	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				log.Println("context cancelled")
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read binance:", err)
					return
				}
				var tickerData BinanceTicker
				err = json.Unmarshal(message, &tickerData)
				if err != nil {
					log.Println("decode error binance", err)
					continue
				}

				// Validate the order book data
				if tickerData.Symbol == "" {
					log.Println("empty symbol received")
					continue
				}

				// Get the best ask price
				if tickerData.AskPrice != "" && tickerData.AskQty != "" {
					symbol := utils.GetQuote(tickerData.Symbol, trie)
					if symbol == "" {
						continue
					}
					tickers <- TickerGeneral{
						Price:  tickerData.AskPrice,
						InstId: symbol,
						Market: "BINANCE",
						Size:   tickerData.AskQty,
					}
				} else {
					log.Printf("invalid Binance ask data for symbol %s", tickerData.Symbol)
					log.Printf("full message: %s", string(message))
				}
			}
		}
	}()

	<-ctx.Done()
	log.Println("binance context done")
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	} else {
		log.Println("write binance close success")
	}
}
