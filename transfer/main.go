package transfer

import (
	"log"
    "os"

    "github.com/joho/godotenv"
)

var (
	binanceAPIKey       string
	binanceAPISecret    string
	kucoinAPIKey        string
	kucoinAPISecret     string
	kucoinPassphrase    string
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

	binanceAPIKey = os.Getenv("BINANCE_API_KEY")
	binanceAPISecret = os.Getenv("BINANCE_API_SECRET")
	kucoinAPIKey = os.Getenv("KUCOIN_API_KEY")
	kucoinAPISecret = os.Getenv("KUCOIN_API_SECRET")
	kucoinPassphrase = os.Getenv("KUCOIN_API_PASSPHRASE")

	// Check if any required environment variable is missing
	if kucoinAPIKey == "" || kucoinAPISecret == "" || kucoinPassphrase == "" || binanceAPIKey == "" || binanceAPISecret == "" {
		log.Fatal("Missing required API keys or secrets in environment variables")
	}
}