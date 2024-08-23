package main

import (
	"context"
	"encoding/json"
	"arbitrage/utils"
	"log"
	"github.com/gorilla/websocket"
)

// Define structures
type BinanceTicker struct {
	Symbol    string `json:"s"`
	Price     string `json:"c"`
}



func binance(ctx context.Context, tickers chan TickerGeneral) {

    trie := utils.Initialize()
	// Binance WebSocket endpoint for ticker data
	endpoint := "wss://stream.binance.com:9443/ws/!miniTicker@arr"

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
				var tickersData []BinanceTicker
				err = json.Unmarshal(message, &tickersData)
				if err != nil {
					log.Println("decode error binance", err)
				}
				for _, ticker := range tickersData {
					symbol := utils.GetQuote(ticker.Symbol, trie)
					if symbol == "" {
						continue
					}
					tickers <- TickerGeneral{
						Price:  ticker.Price,
						InstId: symbol,
						Market: "BINANCE",
					}
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
