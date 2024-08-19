package balance

import (
	"log"
	"strconv"
	"github.com/Kucoin/kucoin-go-sdk"
)

// AccountBalance represents the balance of a specific account.
type AccountBalance struct {
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance,string"`
	Wallet    string  `json:"wallet"`
}

// Helper function to get account balances from a specific endpoint.
func Kucoin()([]AccountBalance, error) {
	s := kucoin.NewApiService(
		kucoin.ApiKeyOption(kucoinAPIKey),
		kucoin.ApiSecretOption(kucoinAPISecret),
		kucoin.ApiPassPhraseOption(kucoinPassphrase),
	)
	rsp, err := s.Accounts("", "")
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return []AccountBalance{}, err
	}

	as := kucoin.AccountsModel{}
	if err := rsp.ReadData(&as); err != nil {
		log.Printf("Error: %s", err.Error())
		return []AccountBalance{}, err
	}

	balances := []AccountBalance{}

	for _, a := range as {
		balances = append(balances, AccountBalance{
			Currency: a.Currency,
			Balance:  func() float64 {
				balance, _ := strconv.ParseFloat(a.Available, 64)
				return balance
			}(),
			Wallet: func () string {
				switch a.Type {
				case "trade":
					return "SPOT"
				case "margin":
					return "MARGIN"
				case "main":
					return "FUNDING"
				default:
					return "UNKNOWN"
				}
			}(),
		})
	}

	return balances, nil
}
