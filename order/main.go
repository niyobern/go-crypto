package order

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Order struct {
	Symbol   string `json:"symbol"`
	OrderID  int64  `json:"orderId"`
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Side     string `json:"side"`
	Time     int64  `json:"time"`
}

var (
	binanceAPIKey    string
	binanceAPISecret string
	kucoinAPIKey     string
	kucoinAPISecret  string
	kucoinPassphrase string
)

type DepositAdress struct {
	Adress string
	Memo   string
	Chain  string
}

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load Binance credentials
	binanceAPIKey = os.Getenv("BINANCE_API_KEY")
	binanceAPISecret = os.Getenv("BINANCE_API_SECRET")

	// Load KuCoin credentials
	kucoinAPIKey = os.Getenv("KUCOIN_API_KEY")
	kucoinAPISecret = os.Getenv("KUCOIN_API_SECRET")
	kucoinPassphrase = os.Getenv("KUCOIN_API_PASSPHRASE")

	// Check if any required environment variable is missing
	if kucoinAPIKey == "" || kucoinAPISecret == "" || kucoinPassphrase == "" ||
		binanceAPIKey == "" || binanceAPISecret == "" {
		log.Fatal("Missing required API keys or secrets in environment variables")
	}
}
