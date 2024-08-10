package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type LoginArgs struct {
	APIKey     string `json:"apiKey"`
	Passphrase string `json:"passphrase"`
	Timestamp  string `json:"timestamp"`
	Sign       string `json:"sign"`
}

type LoginRequest struct {
	Op   string      `json:"op"`
	Args []LoginArgs `json:"args"`
}

type Subscription struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId"`
}

type SubscribeRequest struct {
	Op   string         `json:"op"`
	Args []Subscription `json:"args"`
}

type TickerData struct {
    InstType   string `json:"instType"`
    InstId     string `json:"instId"`
    Last       string `json:"last"`
    LastSz     string `json:"lastSz"`
    AskPx      string `json:"askPx"`
    AskSz      string `json:"askSz"`
    BidPx      string `json:"bidPx"`
    BidSz      string `json:"bidSz"`
    Open24h    string `json:"open24h"`
    High24h    string `json:"high24h"`
    Low24h     string `json:"low24h"`
    SodUtc0    string `json:"sodUtc0"`
    SodUtc8    string `json:"sodUtc8"`
    VolCcy24h  string `json:"volCcy24h"`
    Vol24h     string `json:"vol24h"`
    Ts         string `json:"ts"`
}

type Message struct {
    Arg  struct {
        Channel string `json:"channel"`
        InstId  string `json:"instId"`
    } `json:"arg"`
    Data []TickerData `json:"data"`
}

func getSign(method, requestPath string) (string, string) {
	t := time.Now().Unix()
	timestamp := fmt.Sprintf("%d", t)
	prehashString := timestamp + method + requestPath
	secretKey := "301CC5DAD98447EBB610357C1E8BF2D2"
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(prehashString))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return sign, timestamp
}

// Function to connect to OKX WebSocket and handle messages
func okx(ctx context.Context, tickers chan TickerGeneral) {
	u := url.URL{Scheme: "wss", Host: "ws.okx.com:8443", Path: "/ws/v5/public"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
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
					log.Println("read okx:", err)
					return
				}
				var msg Message
				err = json.Unmarshal([]byte(message), &msg)
				if err != nil {
					log.Fatalf("Error parsing JSON: %v", err)
				}

				if len(msg.Data) > 0 {
					tickers <- TickerGeneral{
						InstId: msg.Data[0].InstId,
						Price:  msg.Data[0].Last,
						Market: "OKX",
					}
				}
			}
		}
	}()

	sign, timestamp := getSign("GET", "/users/self/verify")
	loginArgs := LoginArgs{
		APIKey:     "fe5ae325-5d72-441f-848d-7ee8abd495a9",
		Passphrase: "Reform@781",
		Timestamp:  timestamp,
		Sign:       sign,
	}

	loginRequest := LoginRequest{
		Op:   "login",
		Args: []LoginArgs{loginArgs},
	}
	jsonData, err := json.Marshal(loginRequest)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	err = c.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Println("write:", err)
		return
	}

	subscribeRequest := SubscribeRequest{
		Op: "subscribe",
		Args: []Subscription{
			{Channel: "tickers", InstId: "BTC-USDT"},
			{Channel: "tickers", InstId: "ETH-USDT"},
			{Channel: "tickers", InstId: "OKB-USDT"},
			{Channel: "tickers", InstId: "MATIC-USDT"},
			{Channel: "tickers", InstId: "XRP-USDT"},
			{Channel: "tickers", InstId: "SOL-USDT"},
			{Channel: "tickers", InstId: "DOGE-USDT"},
			{Channel: "tickers", InstId: "PEPE-USDT"},
			{Channel: "tickers", InstId: "SATS-USDT"},
			{Channel: "tickers", InstId: "NOT-USDT"},
			{Channel: "tickers", InstId: "ONDO-USDT"},
			{Channel: "tickers", InstId: "1INCH-USDT"},
			{Channel: "tickers", InstId: "AAVE-USDT"},
			{Channel: "tickers", InstId: "ACA-USDT"},
			{Channel: "tickers", InstId: "ACH-USDT"},
			{Channel: "tickers", InstId: "ADA-USDT"},
			{Channel: "tickers", InstId: "AERGO-USDT"},
			{Channel: "tickers", InstId: "AEVO-USDT"},
			{Channel: "tickers", InstId: "AGLD-USDT"},
			{Channel: "tickers", InstId: "AIDOGE-USDT"},
			{Channel: "tickers", InstId: "AKITA-USDT"},
			{Channel: "tickers", InstId: "ALCX-USDT"},
			{Channel: "tickers", InstId: "ALGO-USDT"},
			{Channel: "tickers", InstId: "ALPHA-USDT"},
			{Channel: "tickers", InstId: "APE-USDT"},
			{Channel: "tickers", InstId: "API3-USDT"},
			{Channel: "tickers", InstId: "APT-USDT"},
			{Channel: "tickers", InstId: "AR-USDT"},
			{Channel: "tickers", InstId: "ARB-USDT"},
			{Channel: "tickers", InstId: "ARG-USDT"},
			{Channel: "tickers", InstId: "ARTY-USDT"},
			{Channel: "tickers", InstId: "AST-USDT"},
			{Channel: "tickers", InstId: "ASTR-USDT"},
			{Channel: "tickers", InstId: "ATH-USDT"},
			{Channel: "tickers", InstId: "ATOM-USDT"},
			{Channel: "tickers", InstId: "AUCTION-USDT"},
			{Channel: "tickers", InstId: "AVAX-USDT"},
			{Channel: "tickers", InstId: "AVIVE-USDT"},
			{Channel: "tickers", InstId: "AXS-USDT"},
			{Channel: "tickers", InstId: "BABYDOGE-USDT"},
			{Channel: "tickers", InstId: "BADGER-USDT"},
			{Channel: "tickers", InstId: "BAL-USDT"},
			{Channel: "tickers", InstId: "BAND-USDT"},
			{Channel: "tickers", InstId: "BAT-USDT"},
			{Channel: "tickers", InstId: "BCH-USDT"},
			{Channel: "tickers", InstId: "BETH-USDT"},
			{Channel: "tickers", InstId: "BICO-USDT"},
			{Channel: "tickers", InstId: "BIGTIME-USDT"},
			{Channel: "tickers", InstId: "BLOK-USDT"},
			{Channel: "tickers", InstId: "BLOCK-USDT"},
			{Channel: "tickers", InstId: "BLUR-USDT"},
			{Channel: "tickers", InstId: "BNB-USDT"},
			{Channel: "tickers", InstId: "BNT-USDT"},
			{Channel: "tickers", InstId: "BONE-USDT"},
			{Channel: "tickers", InstId: "BONK-USDT"},
			{Channel: "tickers", InstId: "BORING-USDT"},
			{Channel: "tickers", InstId: "BORA-USDT"},
			{Channel: "tickers", InstId: "BRWL-USDT"},
			{Channel: "tickers", InstId: "BSV-USDT"},
			{Channel: "tickers", InstId: "BTT-USDT"},
			{Channel: "tickers", InstId: "BZZ-USDT"},
			{Channel: "tickers", InstId: "CEEK-USDT"},
			{Channel: "tickers", InstId: "CELO-USDT"},
			{Channel: "tickers", InstId: "CELR-USDT"},
			{Channel: "tickers", InstId: "CETUS-USDT"},
			{Channel: "tickers", InstId: "CFG-USDT"},
			{Channel: "tickers", InstId: "CFX-USDT"},
			{Channel: "tickers", InstId: "CHZ-USDT"},
			{Channel: "tickers", InstId: "CITY-USDT"},
			{Channel: "tickers", InstId: "CLV-USDT"},
			{Channel: "tickers", InstId: "COMP-USDT"},
			{Channel: "tickers", InstId: "CONV-USDT"},
			{Channel: "tickers", InstId: "CORE-USDT"},
			{Channel: "tickers", InstId: "CRO-USDT"},
			{Channel: "tickers", InstId: "CRV-USDT"},
			{Channel: "tickers", InstId: "CSPR-USDT"},
			{Channel: "tickers", InstId: "CTC-USDT"},
			{Channel: "tickers", InstId: "CTXC-USDT"},
			{Channel: "tickers", InstId: "CVC-USDT"},
			{Channel: "tickers", InstId: "CVX-USDT"},
			{Channel: "tickers", InstId: "CXT-USDT"},
			{Channel: "tickers", InstId: "DAI-USDT"},
			{Channel: "tickers", InstId: "DAO-USDT"},
			{Channel: "tickers", InstId: "DEGEN-USDT"},
			{Channel: "tickers", InstId: "DEP-USDT"},
			{Channel: "tickers", InstId: "DGB-USDT"},
			{Channel: "tickers", InstId: "DIA-USDT"},
			{Channel: "tickers", InstId: "DMAIL-USDT"},
			{Channel: "tickers", InstId: "DORA-USDT"},
			{Channel: "tickers", InstId: "DOT-USDT"},
			{Channel: "tickers", InstId: "DYDX-USDT"},
			{Channel: "tickers", InstId: "EGLD-USDT"},
			{Channel: "tickers", InstId: "ELF-USDT"},
			{Channel: "tickers", InstId: "ELON-USDT"},
			{Channel: "tickers", InstId: "ENJ-USDT"},
			{Channel: "tickers", InstId: "ENS-USDT"},
		},
	}
	subscribeJson, err := json.Marshal(subscribeRequest)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	err = c.WriteMessage(websocket.TextMessage, subscribeJson)
	if err != nil {
		fmt.Println("write subscribe:", err)
		return
	}

	<-ctx.Done()
	log.Println("okx context done")
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	} else {
		log.Println("write okx connection closed")
	}
}
