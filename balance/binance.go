package balance

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/adshao/go-binance/v2"
)

func Binance() ([]AccountBalance, error) {
	// Set up the Binance client with your API key and secret
	client := binance.NewClient("JnVmPqBo5bSZX6NQP5EMJTukpmgcPAJdrirAFGWTfCuJfIeHqHbcrbMIfdiOWpVR", "XKFk5XfmKrTdw4nA8ubFiufyij6uMK1EHDzmNcxALn9dSscaz7kdh6aa0fygUqFl")

	// Get account information including balances
	accountInfo, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error fetching Binance account info: %v", err)
	}
    
	var balances []AccountBalance
	for _, balance := range accountInfo.Balances {
		free, err := strconv.ParseFloat(balance.Free, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		if balance.Asset == "USDT" {
			fmt.Println(balance)
		}

		if free != 0 { // Only display assets with non-zero balance
			amount, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				log.Println(err)
				continue
			}
			bal := AccountBalance{
				Currency: balance.Asset,
				Balance: amount,
			}
			balances = append(balances, bal)
		}
	}
	return balances, nil
}
