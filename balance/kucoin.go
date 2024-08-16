package balance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// KuCoin API credentials (ensure you fill these in)
const (
	baseURL       = "https://api.kucoin.com"
)

// AccountBalance represents the balance of a specific account.
type AccountBalance struct {
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance,string"`
}

// GetSpotAccountBalance retrieves the balance of the spot account on KuCoin.
func KucoinSpot() ([]AccountBalance, error) {
	endpoint := "/api/v1/accounts?type=trade"
	return getAccountBalance(endpoint)
}

// GetMarginAccountBalance retrieves the balance of the margin account on KuCoin.
func KucoinMargin() ([]AccountBalance, error) {
	endpoint := "/api/v1/margin/account"
	return getAccountBalance(endpoint)
}

// Helper function to get account balances from a specific endpoint.
func getAccountBalance(endpoint string) ([]AccountBalance, error) {
	url := baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := signRequest(kucoinAPISecret, timestamp, "GET", endpoint, "")
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
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	// Check for errors in the response
	if resp.StatusCode != http.StatusOK || result["code"] != "200000" {
		return nil, fmt.Errorf("error response: %s", string(body))
	}

	// Extract and return balances
	data := result["data"]
	if endpoint == "/api/v1/margin/account" {
		data = data.(map[string]interface{})["accounts"]
	}

	var balances []AccountBalance
	for _, account := range data.([]interface{}) {
		acc := account.(map[string]interface{})
		balance := AccountBalance{
			Currency:  acc["currency"].(string),
			Balance: parseStringToFloat(acc["availableBalance"].(string)),
		}
		if balance.Balance == 0 {
			continue
		}
		balances = append(balances, balance)
	}
	fmt.Println(balances)

	return balances, nil
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

func parseStringToFloat(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return result
}
