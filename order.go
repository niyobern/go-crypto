package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// KuCoin API credentials
const (
	apiKey        = "your_api_key"
	apiSecret     = "your_api_secret"
	apiPassphrase = "your_api_passphrase"
	baseURL       = "https://api.kucoin.com"
)

func main() {
	// Define the order parameters
	symbol := "SOL-USDT"             // Trading pair
	side := "sell"                   // Buy or sell
	price := "current_market_price"  // Use the current market price, you may set this dynamically
	size := "1000"                   // Amount of SOL to sell
	autoBorrow := true               // Enable auto-borrowing for leverage

	// Create the order
	orderID, err := createMarginOrder(symbol, side, price, size, autoBorrow)
	if err != nil {
		fmt.Printf("Error creating margin order: %v\n", err)
	} else {
		fmt.Printf("Order created successfully with ID: %s\n", orderID)
	}
}

func createMarginOrder(symbol, side, price, size string, autoBorrow bool) (string, error) {
	// Prepare the request payload
	order := map[string]interface{}{
		"symbol":        symbol,
		"side":          side,
		"type":          "limit",  // Use "limit" order type, change to "market" if necessary
		"price":         price,
		"size":          size,
		"autoBorrow":    autoBorrow,
		"tradeType":     "MARGIN_TRADE",  // Specify margin trade
	}

	payload, err := json.Marshal(order)
	if err != nil {
		return "", err
	}

	// Create the request
	endpoint := "/api/v1/margin/order"
	url := baseURL + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	// Add headers
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := signRequest(apiSecret, timestamp, "POST", endpoint, string(payload))
	req.Header.Set("KC-API-KEY", apiKey)
	req.Header.Set("KC-API-SIGN", signature)
	req.Header.Set("KC-API-TIMESTAMP", timestamp)
	req.Header.Set("KC-API-PASSPHRASE", signPassphrase(apiPassphrase))
	req.Header.Set("KC-API-KEY-VERSION", "2")
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Check for errors in the response
	if resp.StatusCode != http.StatusOK || result["code"] != "200000" {
		return "", fmt.Errorf("error response: %s", string(body))
	}

	// Return the order ID
	data := result["data"].(map[string]interface{})
	return data["orderId"].(string), nil
}

func signRequest(secret, timestamp, method, endpoint, body string) string {
	// Prepare the prehash string
	prehash := timestamp + method + endpoint + body

	// Create the HMAC signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(prehash))
	return hex.EncodeToString(h.Sum(nil))
}

func signPassphrase(passphrase string) string {
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(passphrase))
	return hex.EncodeToString(h.Sum(nil))
}
