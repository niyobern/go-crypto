package balance

import (
	"context"
	"log"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"io"
    "encoding/json"
	"net/url"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/adshao/go-binance/v2"
)

func Binance() ([]AccountBalance, error) {
	client := binance.NewClient(binanceAPIKey, binanceAPISecret)

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

		if free != 0 { // Only display assets with non-zero balance
			amount, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				log.Println(err)
				continue
			}
			bal := AccountBalance{
				Currency: balance.Asset,
				Balance: amount,
				Wallet: "SPOT",
			}
			balances = append(balances, bal)
		}
	}
	funding, err := binanceFunding()
	if err != nil {
		return nil, err
	}
    margin, err := binanceMargin()
	if err != nil {
		return nil, err
	}
	balances = append(balances, margin...)
	return append(balances, funding...), nil
}
func binanceMargin() ([]AccountBalance, error) {
	client := binance.NewClient(binanceAPIKey, binanceAPISecret)

	// Get account information including balances
	accountInfo, err := client.NewGetMarginAccountService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error fetching Binance account info: %v", err)
	}

	var balances []AccountBalance
	for _, balance := range accountInfo.UserAssets {
		free, err := strconv.ParseFloat(balance.Free, 64)
		if err != nil {
			log.Println(err)
			continue
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
				Wallet: "MARGIN",
			}
			balances = append(balances, bal)
		}
	}
	return balances, nil
}

func binanceFunding() ([]AccountBalance, error) {
	endpoint := "https://api.binance.com/sapi/v1/asset/get-funding-asset"

	// Set parameters
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{}
	params.Set("timestamp", timestamp)

	// Generate signature
	signature := GenerateSignature(params.Encode(), binanceAPISecret)
	params.Set("signature", signature)

	// Create the request
	req, err := http.NewRequest("POST", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("X-MBX-APIKEY", binanceAPIKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Handle errors from Binance API
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance funding API request failed: %s", string(body))
	}

	// Parse the response
	type BinanceBalance struct{
		Currency  string  `json:"asset"`
		Balance   float64 `json:"free,string"`
		Wallet    string  `json:"wallet"`
	}
	var bals []BinanceBalance
	if err := json.Unmarshal(body, &bals); err != nil {
		return nil, err
	}
	var balances []AccountBalance = make([]AccountBalance, len(bals))
	for i, bal := range bals {
		bal.Wallet = "FUNDING"
		balances[i] = AccountBalance(bal)
	}

	return balances, nil
}


func GenerateSignature(queryString, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}