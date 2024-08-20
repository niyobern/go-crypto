package utils

import (
	"arbitrage/balance"
	"arbitrage/transfer"
	"log"
	"time" 
)

func PostBinanceBuy(base string, capital float64){
	time.Sleep(20 * time.Second)
	bal, err := balance.Binance()
	if err != nil {
		log.Println("Returning Binance Balances, failed")
		return
	}
	for _, coin := range bal {
		if coin.Wallet == "FUNDING" && coin.Currency == base {
			_, err := transfer.Binance2Kucoin(coin.Currency, coin.Balance)
			if err != nil {
				log.Println("Transfer B2K failed", err)
			}
		}
	}
	time.Sleep(60 * time.Second)
	stuckAt := 0 //Helps not to redo the completed step if we the loop continues
	for {
		time.Sleep(10 * time.Second)
		bal, err := balance.Kucoin()
		if err != nil {
			log.Println("Failed to check transfer status:", err)
			continue
		}
		for _, coin := range bal {
			if coin.Wallet == "FUNDING" && coin.Currency == base {
				if stuckAt == 0 {
					_, err := transfer.KucoinFunding2spot(coin.Currency, coin.Balance)
					if err != nil {
						log.Println("Transfer failed KF2S", err)
						continue
					}
					time.Sleep(2 * time.Second)
				}
				if stuckAt <= 1 {
					_, err := transfer.KucoinSpot2Margin(coin.Currency, coin.Balance)
					if err != nil {
						log.Println("Transfer failed KS2M", err)
						stuckAt = 1
						continue
					}
					time.Sleep(2 * time.Second)
				}
				if stuckAt <= 2 {
					_, err := transfer.KucoinRepayLoan(coin.Currency, coin.Balance)
					if err != nil {
						log.Println("Transfer failed KRepay", err)
						stuckAt = 2
						continue
					}
					time.Sleep(2 * time.Second)
				}
				if stuckAt <= 3 {
					_, err := transfer.KucoinMargin2spot("USDT", capital)
					if err != nil {
						log.Println("Transfer failed KM2S", err)
						stuckAt = 3
						continue
					}
					time.Sleep(2 * time.Second)
				}
				if stuckAt <= 4 {
					_, err := transfer.TransferFromKucoinToBinance("USDT", capital)
					if err != nil {
						log.Println("Transfer failed K2B", err)
						stuckAt = 4
						continue
					}
					time.Sleep(2 * time.Second)
				}
				_, err := transfer.BinanceFunding2Spot("USDT", capital)
				if err != nil {
					log.Println("Transfer Failed, BF2S", err)
					stuckAt = 5
				}

				// Transfer from binance funding to spot
				for {
					time.Sleep(30 * time.Second)
					bal, err := balance.Binance()
					if err != nil {
						log.Println("Failed to check transfer status:", err)
						continue
					}
					for _, coin := range bal {
						if coin.Wallet == "FUNDING" && coin.Currency == base {
							if stuckAt == 0 {
								_, err := transfer.BinanceFunding2Spot(coin.Currency, coin.Balance)
								if err != nil {
									log.Println("Transfer failed BF2S", err)
									continue
								}
								return
							}
						}
					}
				}
			}
		}
	}
}


func PostKucoinBuy(base string, capital float64){
	time.Sleep(20 * time.Second)
	bal, err := balance.Kucoin()
	if err != nil {
		log.Println("Returning Binance Balances, failed")
		return
	}
	for _, coin := range bal {
		if coin.Wallet == "FUNDING" && coin.Currency == base {
			_, err := transfer.TransferFromKucoinToBinance(coin.Currency, coin.Balance)
			if err != nil {
				log.Println("Transfer K2B failed", err)
			}
		}
	}
	time.Sleep(60 * time.Second)
	stuckAt := 0 //Helps not to redo the completed step if we the loop continues
	for {
		time.Sleep(10 * time.Second)
		bal, err := balance.Binance()
		if err != nil {
			log.Println("Failed to check transfer status:", err)
			continue
		}
		for _, coin := range bal {
			if coin.Wallet == "FUNDING" && coin.Currency == base {
				if stuckAt == 0 {
					_, err := transfer.BinanceFunding2Spot(coin.Currency, coin.Balance)
					if err != nil {
						log.Println("Transfer failed BF2S", err)
						continue
					}
					time.Sleep(2 * time.Second)
				}
				if stuckAt <= 1 {
					_, err := transfer.BinanceSpot2Margin(coin.Currency, coin.Balance)
					if err != nil {
						log.Println("Transfer failed BS2M", err)
						stuckAt = 1
						continue
					}
					time.Sleep(2 * time.Second)
				}
				if stuckAt <= 2 {
					_, err := transfer.BinanceRepayMarginLoan(coin.Currency, coin.Balance)
					if err != nil {
						log.Println("Transfer failed BRepay", err)
						stuckAt = 2
						continue
					}
					time.Sleep(2 * time.Second)
				}
				if stuckAt <= 3 {
					_, err := transfer.BinanceMargin2Spot("USDT", capital)
					if err != nil {
						log.Println("Transfer failed KM2S", err)
						stuckAt = 3
						continue
					}
					time.Sleep(2 * time.Second)
				}
				if stuckAt <= 4 {
					_, err := transfer.Binance2Kucoin("USDT", capital)
					if err != nil {
						log.Println("Transfer failed B2C", err)
						stuckAt = 4
						continue
					}
					time.Sleep(2 * time.Second)
				}
				_, err := transfer.KucoinFunding2spot("USDT", capital)
				if err != nil {
					log.Println("Transfer Failed, KF2S", err)
					stuckAt = 5
				}

				// Transfer from binance funding to spot
				for {
					time.Sleep(30 * time.Second)
					bal, err := balance.Kucoin()
					if err != nil {
						log.Println("Failed to check transfer status:", err)
						continue
					}
					for _, coin := range bal {
						if coin.Wallet == "FUNDING" && coin.Currency == base {
							if stuckAt == 0 {
								_, err := transfer.KucoinFunding2spot(coin.Currency, coin.Balance)
								if err != nil {
									log.Println("Transfer failed KF2S", err)
									continue
								}
								return
							}
						}
					}
				}
			}
		}
	}
}