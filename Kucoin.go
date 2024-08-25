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

	ch1 := kucoin.NewSubscribeMessage("/market/ticker:JTO-USDT,RUNE-USDT,GRT-USDT,ADA-USDT,SNX-USDT,THETA-USDT,SEI-USDT,ARB-USDT,ENJ-USDT,STORJ-USDT,FLOKI-USDT,SHIB-USDT,JASMY-USDT,ICP-USDT,ENA-USDT,KSM-USDT,FIL-USDT,ILV-USDT,MKR-USDT,CRV-USDT,MANTA-USDT,POND-USDT,SLP-USDT,AXS-USDT,ALGO-USDT,TLM-USDT,LRC-USDT,STRK-USDT,AR-USDT,ERN-USDT,DOT-USDT,CKB-USDT,CAKE-USDT,XLM-USDT,DODO-USDT,SAND-USDT,C98-USDT,APT-USDT,BAT-USDT,FXS-USDT,SUPER-USDT,NKN-USDT,KDA-USDT,ATOM-USDT,AVAX-USDT,APE-USDT,XTZ-USDT,ETH-USDT,FLUX-USDT,BONK-USDT,DYDX-USDT,DYM-USDT,MATIC-USDT,1INCH-USDT,REN-USDT,REQ-USDT,ZIL-USDT,YGG-USDT,WIN-USDT,BB-USDT,NOT-USDT,ARPA-USDT,CELO-USDT,BCH-USDT,DASH-USDT,AEVO-USDT,XRP-USDT,ENS-USDT,WIF-USDT,JUP-USDT,MOVR-USDT,ZRO-USDT,LUNA-USDT,YFI-USDT,USDC-USDT,COMP-USDT,SOL-USDT,ORN-USDT,OGN-USDT,WLD-USDT,CFX-USDT,RSR-USDT,BTC-USDT,EOS-USDT,BLUR-USDT,LPT-USDT,VET-USDT,SKL-USDT,GMT-USDT,ALICE-USDT,FTM-USDT,STX-USDT,ORDI-USDT,ANKR-USDT,UMA-USDT,IOST-USDT,TON-USDT,KAVA-USDT,QNT-USDT,QI-USDT,LTO-USDT,MASK-USDT,BNB-USDT,SUI-USDT,BOME-USDT,LUNC-USDT,API3-USDT,RLC-USDT,ETC-USDT,OP-USDT,PYTH-USDT,UNI-USDT,NEAR-USDT,DGB-USDT,GLMR-USDT,CHR-USDT,LINK-USDT,HBAR-USDT,NEO-USDT,LTC-USDT,MEME-USDT,INJ-USDT,PYR-USDT,SUSHI-USDT,IOTX-USDT,TIA-USDT,DOGE-USDT,WOO-USDT,ROSE-USDT,EGLD-USDT,METIS-USDT,CTSI-USDT,AAVE-USDT,TRX-USDT,CHZ-USDT,MANA-USDT,AUDIO-USDT,PEPE-USDT,LINA-USDT,IMX-USDT,FLOW-USDT,FET-USDT,PEOPLE-USDT,LISTA-USDT,ONE-USDT,CLV-USDT,LDO-USDT,ZEC-USDT,SXP-USDT", false)

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