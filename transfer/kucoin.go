package transfer

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

// KuCoin API configuration
var (
	baseURL = "https://api.kucoin.com"
)

// GetDepositAddress retrieves the deposit address for a given cryptocurrency on KuCoin.
func KucoinGetDepositAddress(currency string) (DepositAdress, error) {
	// Create the request
	endpoint := "/api/v3/deposit-addresses"

	url := fmt.Sprintf("%s%s?currency=%s", baseURL, endpoint, currency)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return DepositAdress{}, err
	}

	// Add headers
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := signRequest(kucoinAPISecret, timestamp, "GET", fmt.Sprintf("%s?currency=%s", endpoint, currency), "")
	req.Header.Set("KC-API-KEY", kucoinAPIKey)
	req.Header.Set("KC-API-SIGN", signature)
	req.Header.Set("KC-API-TIMESTAMP", timestamp)
	req.Header.Set("KC-API-PASSPHRASE", signPassphrase(kucoinPassphrase))
	req.Header.Set("KC-API-KEY-VERSION", "3")
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return DepositAdress{}, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DepositAdress{}, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return DepositAdress{}, err
	}

	// Check for errors in the response
	if resp.StatusCode != http.StatusOK || result["code"] != "200000" {
		return DepositAdress{}, fmt.Errorf("error response: %s", string(body))
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		return DepositAdress{}, nil
	}

	adress := DepositAdress{
		Adress: data[0].(map[string]interface{})["address"].(string),
		Memo:   data[0].(map[string]interface{})["memo"].(string),
		Chain:  data[0].(map[string]interface{})["chainName"].(string),
	}
	return adress, nil
}

func kucoinCreateDepositAdress(currency string) (DepositAdress, error) {
	// Create the request
	endpoint := "/api/v1/deposit-addresses"

	url := fmt.Sprintf("%s%s", baseURL, endpoint)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return DepositAdress{}, err
	}

	// Add headers
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := signRequest(kucoinAPISecret, timestamp, "POST", endpoint, "")
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
		return DepositAdress{}, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DepositAdress{}, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return DepositAdress{}, err
	}

	// Check for errors in the response
	if resp.StatusCode != http.StatusOK || result["code"] != "200000" {
		return DepositAdress{}, fmt.Errorf("error response: %s", string(body))
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		return DepositAdress{}, nil
	}

	adress := DepositAdress{
		Adress: data[0].(map[string]interface{})["address"].(string),
		Memo:   data[0].(map[string]interface{})["memo"].(string),
		Chain:  data[0].(map[string]interface{})["chainName"].(string),
	}
	return adress, nil
}

// TransferSpotToMargin transfers funds from the spot wallet to the margin wallet on KuCoin.
func KucoinSpot2Margin(currency string, amount float64) (string, error) {
	return kucoinInternalTransfer(currency, "trade", "margin", amount)
}

func KucoinMargin2spot(currency string, amount float64) (string, error) {
	return kucoinInternalTransfer(currency, "margin", "trade", amount)
}

func KucoinFunding2spot(currency string, amount float64) (string, error) {
	return kucoinInternalTransfer(currency, "main", "trade", amount)
}

func kucoinInternalTransfer(currency, from, to string, amount float64) (string, error) {
	// Prepare the request payload
	transfer := map[string]interface{}{
		"currency":  currency,
		"amount":    strconv.FormatFloat(amount, 'f', -1, 64),
		"from":      from,
		"to":        to,
		"clientOid": uuid.New().String(),
	}

	jsonBody, err := json.Marshal(transfer)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Create the request
	endpoint := "/api/v2/accounts/inner-transfer"
	url := fmt.Sprintf("%s%s", baseURL, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	// Add headers
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := signRequest(kucoinAPISecret, timestamp, "POST", endpoint, string(jsonBody))
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Check for errors in the response
	if resp.StatusCode != http.StatusOK || result["code"] != "200000" {
		return "", fmt.Errorf("error response: %s", string(body))
	}

	// Return the transfer ID
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		log.Println(result)
		return "", nil
	}
	return data["orderId"].(string), nil
}

func TransferFromKucoinToBinance(currency string, amount float64) (string, error) {
	// Get Binance deposit address
	binanceAdress, err := BinanceDepositAdress(currency, "")
	if err != nil {
		return "", fmt.Errorf("failed to get Binance deposit address: %v", err)
	}

	// Prepare the request payload for the KuCoin withdrawal
	withdrawal := map[string]interface{}{
		"currency":  currency,
		"amount":    strconv.FormatFloat(amount, 'f', -1, 64),
		"address":   binanceAdress.Adress,
		"memo":      binanceAdress.Memo,
		"isInner":   false,
		"clientOid": uuid.New().String(),
		"remark":    "Transfer to Binance",
	}

	jsonBody, err := json.Marshal(withdrawal)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Create the request
	endpoint := "/api/v1/withdrawals"
	url := fmt.Sprintf("%s%s", baseURL, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	// Add headers
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := signRequest(kucoinAPISecret, timestamp, "POST", endpoint, string(jsonBody))
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Check for errors in the response
	if resp.StatusCode != http.StatusOK || result["code"] != "200000" {
		return "", fmt.Errorf("error response: %s", string(body))
	}

	// Return the withdrawal ID
	data := result["data"].(map[string]interface{})
	return data["withdrawalId"].(string), nil
}

func KucoinRepayLoan(currency string, amount float64) (string, error) {
	// Prepare the request payload
	repay := map[string]interface{}{
		"currency":  currency,
		"size":      strconv.FormatFloat(amount, 'f', -1, 64),
		"clientOid": uuid.New().String(),
	}

	jsonBody, err := json.Marshal(repay)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Create the request
	endpoint := "/api/v1/margin/repay"
	url := fmt.Sprintf("%s%s", baseURL, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	// Add headers
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := signRequest(kucoinAPISecret, timestamp, "POST", endpoint, string(jsonBody))
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Check for errors in the response
	if resp.StatusCode != http.StatusOK || result["code"] != "200000" {
		return "", fmt.Errorf("error response: %s", string(body))
	}

	// Return the repayment ID
	data := result["data"].(map[string]interface{})
	return data["orderNo"].(string), nil
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
	h := hmac.New(sha256.New, []byte(kucoinAPISecret))
	h.Write([]byte(passphrase))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
