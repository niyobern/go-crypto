package main

import (
	"log"
    "strings"
	"context"
	"encoding/json"
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

	ch1 := kucoin.NewSubscribeMessage("/market/ticker:SXP-USDT,FLOW-USDT,MANTA-USDT,DYM-USDT,VELO-USDT,BOME-USDT,BB-USDT,ETHW-USDT,HBAR-USDT,LUNA-USDT,EGLD-USDT,APT-USDT,SUI-USDT,QI-USDT,TIA-USDT,WIF-USDT,JTO-USDT,PYTH-USDT,NOT-USDT,NEAR-USDT,OP-USDT,ROSE-USDT,MOVR-USDT,KAVA-USDT,RUNE-USDT,CELO-USDT,SEI-USDT,LISTA-USDT,JUP-USDT,CKB-USDT,KDA-USDT,BONK-USDT,JST-USDT,XEC-USDT,ATEM-USDT,OSMO-USDT,EDU-USDT,ZK-USDT,XYM-USDT,BURGER-USDT,SCRT-USDT,HNT-USDT,NFT-USDT,XAI-USDT,TWT-USDT,IOTA-USDT,VOXEL-USDT,NVT-USDT,STRAX-USDT,AR-USDT,EPX-USDT,XNO-USDT,IO-USDT,GFT-USDT,BSW-USDT,MAGIC-USDT,MBL-USDT,GMX-USDT,TFUEL-USDT,NTRN-USDT,HARD-USDT,FLR-USDT,BNC-USDT,RENDER-USDT,TNSR-USDT,GNS-USDT,TAO-USDT,KLAY-USDT,USTC-USDT,POLYX-USDT,ALPINE-USDT,NFP-USDT", false)

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
				Price:  message.Data.Price,
				InstId: parts[1],
				Market: "KUCOIN",
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