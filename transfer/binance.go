package main

import (
	"context"
	"fmt"
	"log"

	"github.com/adshao/go-binance/v2"
)

func Binance(symbol string, amount float64) {
    // Initialize the Binance client with your API key and secret
    client := binance.NewClient("YOUR_API_KEY", "YOUR_SECRET_KEY")

    // Set the transfer details
    asset := "BTC"        // The asset you want to transfer (e.g., BTC, USDT)

    // Execute the transfer from Spot to Margin
    response, err := client.NewUserUniversalTransferService().
        Asset(asset).
        Amount(amount).
		Type(binance.UserUniversalTransferTypeMainToMargin).
        Do(context.Background())
    
    if err != nil {
        log.Fatalf("Failed to transfer funds: %v", err)
    }

    // Output the result of the transfer
    fmt.Printf("Transfer successful! Transaction ID: %d\n", response.ID)
}
