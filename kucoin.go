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
		"topic":          "/market/ticker:SXP-USDT,FLOW-USDT,MANTA-USDT,DYM-USDT,VELO-USDT,BOME-USDT,BB-USDT,ETHW-USDT,HBAR-USDT,LUNA-USDT,EGLD-USDT,APT-USDT,SUI-USDT,QI-USDT,TIA-USDT,WIF-USDT,JTO-USDT,PYTH-USDT,NOT-USDT,NEAR-USDT,OP-USDT,ROSE-USDT,MOVR-USDT,KAVA-USDT,RUNE-USDT,CELO-USDT,SEI-USDT,LISTA-USDT,JUP-USDT,CKB-USDT,KDA-USDT,BONK-USDT,AR",
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
					log.Println("decode error:", err, message)
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
