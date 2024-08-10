package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// Define structures
type WebSocketTokenResponse struct {
	Code string `json:"code"`
	Data struct {
		Token           string `json:"token"`
		InstanceServers []struct {
			Endpoint     string `json:"endpoint"`
			Protocol     string `json:"protocol"`
			Encrypt      bool   `json:"encrypt"`
			PingInterval int    `json:"pingInterval"`
			PingTimeout  int    `json:"pingTimeout"`
		} `json:"instanceServers"`
	} `json:"data"`
}

type WebSocketMessage struct {
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Data    Data   `json:"data"`
	Subject string `json:"subject"`
}

type Data struct {
	BestAsk     string `json:"bestAsk"`
	BestAskSize string `json:"bestAskSize"`
	BestBid     string `json:"bestBid"`
	BestBidSize string `json:"bestBidSize"`
	Price       string `json:"price"`
	Sequence    string `json:"sequence"`
	Size        string `json:"size"`
	Time        int64  `json:"time"`
}



// Function to get WebSocket token from KuCoin
func getWebSocketToken(apiKey, apiSecret, apiPassphrase string) (string, string, error) {
	client := &http.Client{}
	const kucoinRestURL = "https://api.kucoin.com"

	req, err := http.NewRequest("POST", kucoinRestURL+"/api/v1/bullet-public", nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("KC-API-KEY", apiKey)
	req.Header.Set("KC-API-SECRET", apiSecret)
	req.Header.Set("KC-API-PASSPHRASE", apiPassphrase)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var wsTokenResponse WebSocketTokenResponse
	err = json.Unmarshal(body, &wsTokenResponse)
	if err != nil {
		return "", "", err
	}

	if wsTokenResponse.Code != "200000" {
		return "", "", fmt.Errorf("failed to get WebSocket token: %s", wsTokenResponse.Code)
	}

	return wsTokenResponse.Data.Token, wsTokenResponse.Data.InstanceServers[0].Endpoint, nil
}

// Function to connect to KuCoin WebSocket and handle messages
func kucoin(ctx context.Context, tickers chan TickerGeneral) {
	apiKey := "your_api_key"
	apiSecret := "your_api_secret"
	apiPassphrase := "your_api_passphrase"

	token, endpoint, err := getWebSocketToken(apiKey, apiSecret, apiPassphrase)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the WebSocket
	c, _, err := websocket.DefaultDialer.Dial(endpoint+"?token="+token, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Subscribe to the ticker channel for BTC-USDT and SOL-USDT
	subscribeMsg := map[string]interface{}{
		"id":             "unique_id",
		"type":           "subscribe",
		"topic":          "/market/ticker:BTC-USDT,ETH-USDT,OKB-USDT,MATIC-USDT,XRP-USDT,SOL-USDT,DOGE-USDT,PEPE-USDT,SATS-USDT,NOT-USDT,ONDO-USDT,1INCH-USDT,AAVE-USDT,ACA-USDT,ACH-USDT,ADA-USDT,AERGO-USDT,AEVO-USDT,AGLD-USDT,AIDOGE-USDT,AKITA-USDT,ALCX-USDT,ALGO-USDT,ALPHA-USDT,APE-USDT,API3-USDT,APT-USDT,AR-USDT,ARB-USDT,ARG-USDT,ARTY-USDT,AST-USDT,ASTR-USDT,ATH-USDT,ATOM-USDT,AUCTION-USDT,AVAX-USDT,AVIVE-USDT,AXS-USDT,BABYDOGE-USDT,BADGER-USDT,BAL-USDT,BAND-USDT,BAT-USDT,BCH-USDT,BETH-USDT,BICO-USDT,BIGTIME-USDT,BLOK-USDT,BLOCK-USDT,BLUR-USDT,BNB-USDT,BNT-USDT,BONE-USDT,BONK-USDT,BORING-USDT,BORA-USDT,BRWL-USDT,BSV-USDT,BTT-USDT,BZZ-USDT,CEEK-USDT,CELO-USDT,CELR-USDT,CETUS-USDT,CFG-USDT,CFX-USDT,CHZ-USDT,CITY-USDT,CLV-USDT,COMP-USDT,CONV-USDT,CORE-USDT,CRO-USDT,CRV-USDT,CSPR-USDT,CTC-USDT,CTXC-USDT,CVC-USDT,CVX-USDT,CXT-USDT,DAI-USDT,DAO-USDT,DEGEN-USDT,DEP-USDT,DGB-USDT,DIA-USDT,DMAIL-USDT,DORA-USDT,DOT-USDT,DYDX-USDT,EGLD-USDT,ELF-USDT,ELON-USDT,ENJ-USDT,ENS-USDT",
		// "topic":          "/market/ticker:DGB-USDT,XRP-USDT,XLM-USDT,TRX-USDT,NANO-USDT,SOL-USDT,VET-USDT,ALGO-USDT,MATIC-USDT,EOS-USDT,XTZ-USDT,ATOM-USDT,HBAR-USDT,FTM-USDT,DASH-USDT,ADA-USDT,EGLD-USDT,AVAX-USDT,ZIL-USDT,LUNC-USD",
		"privateChannel": false,
		"response":       true,
	}

	err = c.WriteJSON(subscribeMsg)
	if err != nil {
		log.Fatal("write:", err)
	}

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
					log.Println("read kucoin:", err)
					return
				}
				var msg WebSocketMessage
				err = json.Unmarshal(message, &msg)
				if err != nil {
					log.Fatal("decode error:", err, message)
				}
				parts := strings.Split(msg.Topic, ":")
				if len(parts) < 2 {
					continue
				}
				tickers <- TickerGeneral{
					Price:  msg.Data.Price,
					InstId: parts[1],
					Market: "KUCOIN",
				}
			}
		}
	}()

	<-ctx.Done()
	log.Println("kucoin context done")
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	} else {
        log.Println("write kucoin close success")
    }
}
