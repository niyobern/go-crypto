package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/Kucoin/kucoin-go-sdk"
)

const (
	apiKey        = "66aaab422dc99c0001efe888"
	apiSecret     = "98fe74d7-6915-43b7-b95a-4346ab99ad80"
	apiPassphrase = "Reform@781"
	baseURL       = "https://api.kucoin.com"
)

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

func Kucoin(ctx context.Context, tickers chan TickerGeneral) {
	//s := kucoin.NewApiServiceFromEnv()
	s := kucoin.NewApiService(
		kucoin.ApiKeyOption(apiKey),
		kucoin.ApiSecretOption(apiSecret),
		kucoin.ApiPassPhraseOption(apiPassphrase),
	)
	publicWebsocket(s, tickers)
	<-ctx.Done()
	log.Println("kucoin context done")
}

func publicWebsocket(s *kucoin.ApiService, tickers chan TickerGeneral) {
	rsp, err := s.WebSocketPublicToken()
	if err != nil {
		log.Printf("Error kucoin: %s", err.Error())
		return
	}

	tk := &kucoin.WebSocketTokenModel{}
	if err := rsp.ReadData(tk); err != nil {
		log.Printf("Error kucoin: %s", err.Error())
		return
	}

	c := s.NewWebSocketClient(tk)

	mc, ec, err := c.Connect()
	if err != nil {
		log.Printf("Error kucoin: %s", err.Error())
		return
	}

	// Using the same symbols as Binance
	ch1 := kucoin.NewSubscribeMessage("/market/ticker:CFX-USDT,CELO-USDT,FLOW-USDT,ICP-USDT,ICX-USDT,FIL-USDT,XLM-USDT,ALGO-USDT,ENJ-USDT,DGB-USDT,ONE-USDT,IOST-USDT,RVN-USDT,ZIL-USDT,WAXP-USDT,ACE-USDT,EGLD-USDT,GAS-USDT,HBAR-USDT,EOS-USDT,DYDX-USDT,MOVR-USDT,POL-USDT,WLD-USDT,APT-USDT,IOTA-USDT,DOGE-USDT,XTZ-USDT,NEAR-USDT,LTC-USDT,ATOM-USDT,RDNT-USDT,ARB-USDT,MAGIC-USDT,NEO-USDT,USDC-USDT,TON-USDT,XRP-USDT,ETH-USDT,QTUM-USDT,SUI-USDT,GMT-USDT,CATI-USDT,OP-USDT,ETC-USDT,AR-USDT,HMSTR-USDT,FLM-USDT,AVAX-USDT,THETA-USDT,TRX-USDT,ONT-USDT,KSM-USDT,INJ-USDT,NOT-USDT,GMX-USDT,BCH-USDT,ADA-USDT,DOT-USDT,ZRO-USDT,ELF-USDT,METIS-USDT,MINA-USDT,TIA-USDT,JOE-USDT,DOGE-USDT,STRK-USDT,RENDER-USDT,BNB-USDT,FLOKI-USDT,JST-USDT,TNSR-USDT,W-USDT,JTO-USDT,BONK-USDT,PYTH-USDT,WIF-USDT,JUP-USDT,SOL-USDT,RAY-USDT,BOME-USDT", false)

	if err := c.Subscribe(ch1); err != nil {
		log.Printf("Error kucoin: %s", err.Error())
		return
	}

	for {
		select {
		case err := <-ec:
			c.Stop() // Stop subscribing the WebSocket feed
			log.Printf("Error kucoin: %s", err.Error())
			return
		case msg := <-mc:
			var message WebSocketMessage
			err := json.Unmarshal([]byte(kucoin.ToJsonString(msg)), &message)
			if err != nil {
				log.Println("decode error kucoin:", err, message)
			}
			parts := strings.Split(message.Topic, ":")
			if len(parts) < 2 {
				continue
			}
			tickers <- TickerGeneral{
				Price:  message.Data.BestBid,
				InstId: parts[1],
				Market: "KUCOIN",
				Size:   message.Data.BestBidSize,
			}
		}
	}
}

// func orders(s *kucoin.ApiService) {
// 	rsp, err := s.Orders(map[string]string{}, &kucoin.PaginationParam{CurrentPage: 1, PageSize: 10})
// 	if err != nil {
// 		log.Printf("Error: %s", err.Error())
// 		return
// 	}

// 	os := kucoin.OrdersModel{}
// 	pa, err := rsp.ReadPaginationData(&os)
// 	if err != nil {
// 		log.Printf("Error: %s", err.Error())
// 		return
// 	}
// 	log.Printf("Total num: %d, total page: %d", pa.TotalNum, pa.TotalPage)
// 	for _, o := range os {
// 		log.Printf("Order: %s, %s, %s", o.Id, o.Type, o.Price)
// 	}
// }

// func privateWebsocket(s *kucoin.ApiService) {
// 	rsp, err := s.WebSocketPrivateToken()
// 	if err != nil {
// 		log.Printf("Error: %s", err.Error())
// 		return
// 	}

// 	tk := &kucoin.WebSocketTokenModel{}
// 	//tk.AcceptUserMessage = true
// 	if err := rsp.ReadData(tk); err != nil {
// 		log.Printf("Error: %s", err.Error())
// 		return
// 	}

// 	c := s.NewWebSocketClient(tk)

// 	mc, ec, err := c.Connect()
// 	if err != nil {
// 		log.Printf("Error: %s", err.Error())
// 		return
// 	}

// 	ch1 := kucoin.NewSubscribeMessage("/market/level3:BTC-USDT", false)
// 	ch2 := kucoin.NewSubscribeMessage("/account/balance", false)

// 	if err := c.Subscribe(ch1, ch2); err != nil {
// 		log.Printf("Error: %s", err.Error())
// 		return
// 	}

// 	for {
// 		select {
// 		case err := <-ec:
// 			c.Stop() // Stop subscribing the WebSocket feed
// 			log.Printf("Error: %s", err.Error())
// 			return
// 		case msg := <-mc:
// 			log.Printf("Received: %s", kucoin.ToJsonString(msg))
// 		}
// 	}
// }
