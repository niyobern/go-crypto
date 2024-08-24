package order

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// KuCoin API credentials
const (
	apiKey        = "66aaab422dc99c0001efe888"
	apiSecret     = "98fe74d7-6915-43b7-b95a-4346ab99ad80"
	apiPassphrase = "Reform@781"
	baseURL       = "https://api.kucoin.com"
)

func Kucoin(orderType string, symbol string, amount float64) error {
	if orderType == "SPOT" {
		side := "buy" // Only buying is allowed for spot orders
		size := strconv.FormatFloat(amount, 'f', -1, 64) // Amount to buy

		// Create the spot order
		_, err := createSpotOrder(symbol, side, size)
		if err != nil {
			return err
		} else {
			return nil
		}
	} else if orderType == "MARGIN" {
		side := "sell"                   // Selling only
		size := strconv.FormatFloat(amount, 'f', -1, 64) // Amount to sell
		autoBorrow := true  // Enable auto-borrowing for leverage

		_, err := createMarginOrder(symbol, side, size, autoBorrow)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
	return nil
}

func KucoinReverse(orderType string, symbol string, amount float64) error {
	if orderType == "SPOT" {
		side := "sell" // Only buying is allowed for spot orders
		size := strconv.FormatFloat(amount, 'f', -1, 64) // Amount to buy

		// Create the spot order
		_, err := createSpotOrder(symbol, side, size)
		if err != nil {
			return err
		} else {
			return nil
		}
	} else if orderType == "MARGIN" {
		side := "buy"                   // Selling only
		size := strconv.FormatFloat(amount, 'f', -1, 64) // Amount to sell
		autoBorrow := true  // Enable auto-borrowing for leverage

		// Create the margin order
		_, err := createMarginOrder(symbol, side, size, autoBorrow)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
	return nil
}

func createSpotOrder(symbol, side, size string) (string, error) {
	// Prepare the request payload for spot order
	order := map[string]interface{}{
		"clientOid": uuid.New().String(),
		"symbol":    symbol,
		"side":      side,
		"type":      "market", // Market order type
		"size":      size,
	}

	payload, err := json.Marshal(order)
	if err != nil {
		return "", err
	}

	// Create the request
	endpoint := "/api/v1/orders"
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Println(body)

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

func createMarginOrder(symbol, side, size string, autoBorrow bool) (string, error) {
	// Prepare the request payload for margin order
	order := map[string]interface{}{
		"clientOid":     uuid.New().String(),
		"symbol":        symbol,
		"side":          side,
		"type":          "market",
		"size":          size,
		"marginModel":   "cross",
		"autoBorrow":    autoBorrow,
		"tradeType":     "MARGIN_TRADE",
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Println(body)

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
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func signPassphrase(passphrase string) string {
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(passphrase))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
