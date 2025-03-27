package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func KucoinLimit(orderType string, symbol string, price, amount float64) {
	if orderType == "SPOT" {
		side := "sell"                                   // Only buying is allowed for spot orders
		size := strconv.FormatFloat(amount, 'f', -1, 64) // Amount to buy

		// Create the spot order
		orderID, err := createSpotOrderLimit(symbol, side, size, price)
		if err != nil {
			fmt.Printf("Kucoin spot error: %v\n", err)
		} else {
			fmt.Printf("Spot order created successfully with ID: %s\n", orderID)
		}
	} else if orderType == "MARGIN" {
		side := "buy"                                    // Selling only
		size := strconv.FormatFloat(amount, 'f', -1, 64) // Amount to sell
		autoBorrow := true                               // Enable auto-borrowing for leverage

		// Create the margin order
		orderID, err := createMarginOrderLimit(symbol, side, size, autoBorrow, price)
		if err != nil {
			log.Printf("Kucoin order error: %v\n", err)
		} else {
			fmt.Printf("Margin order created successfully with ID: %s\n", orderID)
		}
	}
}

func createSpotOrderLimit(symbol, side, size string, price float64) (string, error) {
	// Prepare the request payload for spot order
	order := map[string]interface{}{
		"clientOid": uuid.New().String(),
		"symbol":    symbol,
		"side":      side,
		"type":      "limit", // Market order type
		"size":      size,
		"price":     fmt.Sprintf("%f", price),
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
	signature := signRequest(kucoinAPISecret, timestamp, "POST", endpoint, string(payload))
	req.Header.Set("KC-API-KEY", kucoinAPIKey)
	req.Header.Set("KC-API-SIGN", signature)
	req.Header.Set("KC-API-TIMESTAMP", timestamp)
	req.Header.Set("KC-API-PASSPHRASE", signPassphrase(kucoinPassphrase))
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

func createMarginOrderLimit(symbol, side, size string, autoBorrow bool, price float64) (string, error) {
	// Prepare the request payload for margin order
	order := map[string]interface{}{
		"clientOid":   uuid.New().String(),
		"symbol":      symbol,
		"side":        side,
		"type":        "limit",
		"size":        size,
		"price":       fmt.Sprintf("%f", price),
		"marginModel": "cross",
		"autoBorrow":  autoBorrow,
		"tradeType":   "MARGIN_TRADE",
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
	signature := signRequest(kucoinAPISecret, timestamp, "POST", endpoint, string(payload))
	req.Header.Set("KC-API-KEY", kucoinAPIKey)
	req.Header.Set("KC-API-SIGN", signature)
	req.Header.Set("KC-API-TIMESTAMP", timestamp)
	req.Header.Set("KC-API-PASSPHRASE", signPassphrase(kucoinPassphrase))
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
