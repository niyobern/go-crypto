package transfer

import (
	"errors"
	"sort"
    "bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
    "io"
	"math"
    "encoding/json"
	"net/url"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

)


func BinanceSpot2Margin(asset string, amount float64) (string, error){
    return binanceInternal(asset, "MAIN_MARGIN", amount)
}

func BinanceMargin2Spot(asset string, amount float64) (string, error){
    return binanceInternal(asset, "MARGIN_MAIN", amount)
}

func BinanceFunding2Spot(asset string, amount float64) (string, error){
	return binanceInternal(asset, "FUNDING_MAIN", amount)
}

func GenerateSignature(queryString, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

type TransferResponse struct {
	TranID int64 `json:"tranId"`
}
// TransferFunds handles transferring funds from margin account to spot account
func binanceInternal(asset, transferType string, amount float64) (string, error) {
	endpoint := "https://api.binance.com/sapi/v1/asset/transfer"

	// Set parameters
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{}
	params.Set("type", transferType) // Transfer from Margin (cross) to Spot
	params.Set("asset", asset)
	params.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	params.Set("timestamp", timestamp)

	// Generate signature
	signature := GenerateSignature(params.Encode(), binanceAPISecret)
	params.Set("signature", signature)

	// Create the request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("X-MBX-APIKEY", binanceAPIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Handle errors from Binance API
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed: %s", string(body))
	}

	// Parse the response
	var transferResponse TransferResponse
	if err := json.Unmarshal(body, &transferResponse); err != nil {
		return "", err
	}

	return strconv.FormatInt(transferResponse.TranID, 10), nil
}

func wround(stamount string, multiple float64) string {
	log.Println(multiple)
	amount, _ := strconv.ParseFloat(stamount, 64)
	roun := math.Floor(math.Floor(amount/multiple) * multiple)
	log.Println(roun)
	return strconv.FormatFloat(roun, 'f', -1, 64)
}

func withdrawFunds(asset, amount, address, memo, network string) (string, error) {
	endpoint := "https://api.binance.com/sapi/v1/capital/withdraw/apply"

	coins, _ := getAllCoins()
	var multiple float64

	for _, coin := range coins {
		for _, chain := range coin.NetworkList {
			if chain.Network == network {
				multiple, _ = strconv.ParseFloat(chain.WithdrawIntegerMultiple, 64)
			}
		}
	}

	// Set parameters
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{}
	params.Set("coin", asset)
	params.Set("address", address)
	params.Set("amount", wround(amount, multiple))
	params.Set("network", network)
	params.Set("timestamp", timestamp)

	// Optional fields
	if memo != "" {
		params.Set("addressTag", memo) // Memo or AddressTag for some networks
	}

	// Generate signature
	signature := GenerateSignature(params.Encode(), binanceAPISecret)
	params.Set("signature", signature)

	// Create the request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("X-MBX-APIKEY", binanceAPIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Handle errors from Binance API
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed: %s", string(body))
	}

	// Parse the response
	var withdrawResponse WithdrawResponse
	if err := json.Unmarshal(body, &withdrawResponse); err != nil {
		return "", err
	}

	return withdrawResponse.ID, nil
}

type WithdrawResponse struct {
	ID string `json:"id"`
}

type NetworkInfo struct {
	AddressRegex            string `json:"addressRegex"`
	Coin                    string `json:"coin"`
	DepositEnable           bool   `json:"depositEnable"`
	DepositDesc             string `json:"depositDesc,omitempty"`
	IsDefault               bool   `json:"isDefault"`
	MemoRegex               string `json:"memoRegex"`
	MinConfirm              int    `json:"minConfirm"`
	Name                    string `json:"name"`
	Network                 string `json:"network"`
	SpecialTips             string `json:"specialTips,omitempty"`
	UnLockConfirm           int    `json:"unLockConfirm"`
	WithdrawEnable          bool   `json:"withdrawEnable"`
	WithdrawFee             string `json:"withdrawFee"`
	WithdrawIntegerMultiple string `json:"withdrawIntegerMultiple"`
	WithdrawMax             string `json:"withdrawMax"`
	WithdrawMin             string `json:"withdrawMin"`
	SameAddress             bool   `json:"sameAddress"`
	EstimatedArrivalTime    int    `json:"estimatedArrivalTime"`
	Busy                    bool   `json:"busy"`
	ContractAddressUrl      string `json:"contractAddressUrl,omitempty"`
	ContractAddress         string `json:"contractAddress,omitempty"`
}

// CoinInfo represents the structure for each coin's information returned by the API
type CoinInfo struct {
	Coin              string        `json:"coin"`
	DepositAllEnable  bool          `json:"depositAllEnable"`
	Free              string        `json:"free"`
	Freeze            string        `json:"freeze"`
	Ipoable           string        `json:"ipoable"`
	Ipoing            string        `json:"ipoing"`
	IsLegalMoney      bool          `json:"isLegalMoney"`
	Locked            string        `json:"locked"`
	Name              string        `json:"name"`
	NetworkList       []NetworkInfo `json:"networkList"`
	Storage           string        `json:"storage"`
	Trading           bool          `json:"trading"`
	WithdrawAllEnable bool          `json:"withdrawAllEnable"`
	Withdrawing       string        `json:"withdrawing"`
}

type MarginCoin struct {
	Base          string `json:"base"`
	ID            int64  `json:"id"`
	IsBuyAllowed  bool   `json:"isBuyAllowed"`
	IsMarginTrade bool   `json:"isMarginTrade"`
	IsSellAllowed bool   `json:"isSellAllowed"`
	Quote         string `json:"quote"`
	Symbol        string `json:"symbol"`
}

func GetAllMarginAllowed() ([]string, error) {
	endpoint := "https://api.binance.com/sapi/v1/margin/allPairs"

	// Set parameters
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{}
	params.Set("timestamp", timestamp)

	// Generate signature
	signature := GenerateSignature(params.Encode(), binanceAPISecret)
	params.Set("signature", signature)

	// Create the request
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
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
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	// Parse the response
	var coins []MarginCoin
	if err := json.Unmarshal(body, &coins); err != nil {
		return nil, err
	}

	var allowed []string
	for _, i := range coins {
		if i.IsBuyAllowed && i.IsMarginTrade && i.IsSellAllowed && i.Quote == "USDT"{
			allowed = append(allowed, i.Base)
		}
	}

	return allowed, nil
}

// getAllCoins fetches all coins available for deposit and withdrawal
func getAllCoins() ([]CoinInfo, error) {
	endpoint := "https://api.binance.com/sapi/v1/capital/config/getall"

	// Set parameters
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{}
	params.Set("timestamp", timestamp)

	// Generate signature
	signature := GenerateSignature(params.Encode(), binanceAPISecret)
	params.Set("signature", signature)

	// Create the request
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
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
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	// Parse the response
	var coins []CoinInfo
	if err := json.Unmarshal(body, &coins); err != nil {
		return nil, err
	}

	return coins, nil
}

func Binance2Kucoin(asset string, amount float64) (string, error) {

	// Step 1: Get the deposit address from KuCoin
	kucoinAddress, err := KucoinGetDepositAddress(asset)
	if err != nil {
		log.Fatalf("Error getting KuCoin deposit address: %v", err)
	}

	// Step 2: Transfer funds from Binance to the KuCoin deposit address
	withdrawalID, err := withdrawFunds(asset, strconv.FormatFloat(amount, 'f', -1, 64), kucoinAddress.Adress, kucoinAddress.Memo, kucoinAddress.Chain)
	if err != nil {
		log.Fatalf("Error withdrawing funds from Binance: %v", err)
        return "", err
	}

	fmt.Printf("Funds transferred successfully. Withdrawal ID: %s\n", withdrawalID)
    return withdrawalID, nil
}

func getDepositAddress(coin, network string) (*DepositAddressResponse, error) {
	endpoint := "https://api.binance.com/sapi/v1/capital/deposit/address"

	// Set parameters
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{}
	params.Set("coin", coin)
	params.Set("network", network)
	params.Set("timestamp", timestamp)

	// Generate signature
	signature := GenerateSignature(params.Encode(), binanceAPISecret)
	params.Set("signature", signature)

	// Create the request
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
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
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	// Parse the response
	var depositAddress DepositAddressResponse
	if err := json.Unmarshal(body, &depositAddress); err != nil {
		return nil, err
	}

	return &depositAddress, nil
}

type DepositAddressResponse struct {
	Coin     string `json:"coin"`
	Address  string `json:"address"`
	Tag      string `json:"tag,omitempty"`
	URL      string `json:"url"`
}

func BinanceDepositAdress(coin, failedNetwork string) (DepositAdress, error){

    coins, err := getAllCoins()
    if err != nil {
        return DepositAdress{}, err
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

    res, err := getDepositAddress(coin, chain)

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


func BinanceRepayMarginLoan(asset string, amount float64) (string, error) {
	endpoint := "https://api.binance.com/sapi/v1/margin/borrow-repay"

	// Set parameters
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{}
	params.Set("asset", asset)
	params.Set("isIsolated", "FALSE")
	params.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	params.Set("type", "REPAY")
	params.Set("timestamp", timestamp)

	// Generate signature
	signature := GenerateSignature(params.Encode(), binanceAPISecret)
	params.Set("signature", signature)

	// Create the request
	req, err := http.NewRequest("POST", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("X-MBX-APIKEY", binanceAPIKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Handle errors from Binance API
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed: %s", string(body))
	}

	// Parse the response
	type TransactionResponse struct {
		TranID string `json:"tranId"`
	}
	var transaction TransactionResponse
	if err := json.Unmarshal(body, &transaction); err != nil {
		return "", err
	}

	return transaction.TranID, nil
}

