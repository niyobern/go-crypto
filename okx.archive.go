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
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
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
	InstType  string `json:"instType"`
	InstId    string `json:"instId"`
	Last      string `json:"last"`
	LastSz    string `json:"lastSz"`
	AskPx     string `json:"askPx"`
	AskSz     string `json:"askSz"`
	BidPx     string `json:"bidPx"`
	BidSz     string `json:"bidSz"`
	Open24h   string `json:"open24h"`
	High24h   string `json:"high24h"`
	Low24h    string `json:"low24h"`
	SodUtc0   string `json:"sodUtc0"`
	SodUtc8   string `json:"sodUtc8"`
	VolCcy24h string `json:"volCcy24h"`
	Vol24h    string `json:"vol24h"`
	Ts        string `json:"ts"`
}

type Message struct {
	Arg struct {
		Channel string `json:"channel"`
		InstId  string `json:"instId"`
	} `json:"arg"`
	Data []TickerData `json:"data"`
}

func getSign(method, requestPath string) (string, string) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	secretKey := os.Getenv("OKX_API_SECRET")
	if secretKey == "" {
		log.Fatal("Missing required OKX API secret in environment variables")
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	prehashString := timestamp + method + requestPath
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
	symbols := []string{"XEC-USDT", "KAVA-USDT", "EDU-USDT", "AR-USDT", "HNT-USDT", "LUNA-USDT", "BSW-USDT", "BNC-USDT", "FLOW-USDT", "OP-USDT", "LISTA-USDT", "NTRN-USDT", "NVT-USDT", "HARD-USDT", "QI-USDT", "ZK-USDT", "VOXEL-USDT", "DYM-USDT", "TWT-USDT", "MBL-USDT", "ATEM-USDT", "BURGER-USDT", "TFUEL-USDT", "XAI-USDT", "NFP-USDT", "MAGIC-USDT", "CKB-USDT", "MOVR-USDT", "APT-USDT", "SUI-USDT", "RENDER-USDT", "NOT-USDT", "GMX-USDT", "GNS-USDT", "ROSE-USDT", "TIA-USDT", "EPX-USDT", "IO-USDT", "USTC-USDT", "OSMO-USDT", "TAO-USDT", "TNSR-USDT", "BONK-USDT", "SEI-USDT", "IOTA-USDT", "WIF-USDT", "JST-USDT", "ETHW-USDT", "FLR-USDT", "NEAR-USDT", "MANTA-USDT", "RUNE-USDT", "CELO-USDT", "SXP-USDT", "PYTH-USDT", "HBAR-USDT", "KLAY-USDT", "JTO-USDT", "KDA-USDT", "EGLD-USDT", "GFT-USDT", "STRAX-USDT", "VELO-USDT", "XYM-USDT", "NFT-USDT", "BOME-USDT", "JUP-USDT", "SCRT-USDT", "POLYX-USDT", "XNO-USDT", "ALPINE-USDT", "BB-USDT"}
	args := make([]Subscription, 0)
	for _, symbol := range symbols {
		args = append(args, Subscription{Channel: "tickers", InstId: symbol})
	}
	subscribeRequest := SubscribeRequest{
		Op:   "subscribe",
		Args: args,
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
