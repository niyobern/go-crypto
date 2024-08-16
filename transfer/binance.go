package transfer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/adshao/go-binance/v2"
)


func BinanceSpot2Margin(asset string, amount float64) {
    // Initialize the Binance client with your API key and secret
    client := binance.NewClient(binanceAPIKey, binanceAPIKey)

    // Execute the transfer from Margin ro Spot
    response, err := client.NewUserUniversalTransferService().
        Asset(asset).
        Amount(amount).
		Type(binance.UserUniversalTransferTypeMarginToMain).
        Do(context.Background())
    
    if err != nil {
        log.Fatalf("Failed to transfer funds: %v", err)
    }

    // Output the result of the transfer
    fmt.Printf("Transfer successful! Transaction ID: %d\n", response.ID)
}

func BinanceMargin2Spot(asset string, amount float64) {
    // Initialize the Binance client with your API key and secret
    client := binance.NewClient(binanceAPIKey, binanceAPIKey)


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

func withdrawFunds(asset, amount, address, memo, network string) (string, error) {
   
    client := binance.NewClient(binanceAPIKey, binanceAPIKey)
    res, err := client.NewCreateWithdrawService().
        Address(address).
        AddressTag(memo).
        Name(memo).
        Network(network).
        Amount(amount).
        Coin(asset).
        Do(context.Background())

	if err != nil {
		return "", err
	}

	return res.ID, nil
}
func Binance2Kucoin() {
	asset := "USDT"
	amount := "100" // Amount to transfer

	// Step 1: Get the deposit address from KuCoin
	kucoinAddress, err := KucoinGetDepositAddress(asset)
	if err != nil {
		log.Fatalf("Error getting KuCoin deposit address: %v", err)
	}

	// Step 2: Transfer funds from Binance to the KuCoin deposit address
	withdrawalID, err := withdrawFunds(asset, amount, kucoinAddress.Adress, kucoinAddress.Memo, kucoinAddress.Chain)
	if err != nil {
		log.Fatalf("Error withdrawing funds from Binance: %v", err)
	}

	fmt.Printf("Funds transferred successfully. Withdrawal ID: %s\n", withdrawalID)
}

func BinanceDepositAdress(coin, failedNetwork string) (DepositAdress, error){
    client := binance.NewClient(binanceAPIKey, binanceAPIKey)

    coins, err := client.NewGetAllCoinsInfoService().Do(context.Background())
    if err != nil {
        return DepositAdress{}, nil
    }

    var chain string
    
    for _, c := range coins{
        if c.Coin == coin {
            type FeeInfo struct {
                Network string
                Fee      float64
            }
            fees := []FeeInfo{}
            for _, network := range c.NetworkList {
                if network.WithdrawEnable{
                    fee, err := strconv.ParseFloat(network.WithdrawFee, 64)
                    if err != nil {
                        fmt.Println("error converting string to float64:", err)
                        continue
                    }
                    fees = append(fees, FeeInfo{Network: network.Network, Fee: fee})
                }
            }
            sort.Slice(fees, func(i, j int) bool {
                return fees[i].Fee < fees[j].Fee
            })
            findNetworkIndex := func(Network string) int {
                for i, feeInfo := range fees {
                    if feeInfo.Network == Network {
                        return i
                    }
                }
                return -1 // Return -1 if the Network is not found
            }
            index := findNetworkIndex(failedNetwork)
            if index >= len(fees)-1 {
                err := errors.New("no network found")
                return DepositAdress{}, err
            }
            if index == -1 {
                chain = fees[0].Network
            } else {
                chain = fees[index+1].Network
            }
        }
    }

    if chain == ""{
        err := errors.New("no network found")
        return DepositAdress{}, err
    }

    res, err := client.NewGetDepositAddressService().Coin(coin).Do(context.Background())

    if err != nil {
        return DepositAdress{}, err
    }

    adress := DepositAdress{
        Adress: res.Address,
        Memo: res.Tag,
        Chain: chain,
    }

    return adress, nil
}

func BinanceRepayMarginLoan(asset string, amount float64) {
    // Initialize the Binance client with your API key and secret
    client := binance.NewClient(binanceAPIKey, binanceAPIKey)

    // Execute the transfer from Spot to Margin
    response, err := client.NewMarginRepayService().
            Asset(asset).
            Amount(strconv.FormatFloat(amount, 'f', -1, 64)).
            IsIsolated(false).
            Do(context.Background())
    
    if err != nil {
        log.Fatalf("Failed to repay margin loan: %v", err)
    }

    // Output the result of the transfer
    fmt.Printf("Repayment successful! Transaction ID: %d\n", response.TranID)
}
