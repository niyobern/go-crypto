package utils

import (
    "log"
    "os"
    "time"
	"strconv"
	"fmt"

    "github.com/syndtr/goleveldb/leveldb"
    "github.com/syndtr/goleveldb/leveldb/errors"
)

type OrderData struct {
	BuyMarket string
	SellMarket string
	BuyPrice float64
	SellPrice float64
	Amount float64
	Coin string
}

func Database(dbPath string) (*leveldb.DB, error){
	removeLockFile(dbPath)

	// Attempt to open the database
	db, err := openLevelDBWithRetry(dbPath, 3)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func SaveOrders(db *leveldb.DB, buyMarket, sellMarket, coin string, amount, buyPrice, sellPrice float64) error{
	// Save data to the database
	err := db.Put([]byte("buyMarket"), []byte(buyMarket), nil)
	if err != nil {
		return err
	}

	err = db.Put([]byte("coin"), []byte(coin), nil)
	if err != nil {
		return err
	}

	err = db.Put([]byte("minPrice"), []byte(fmt.Sprintf("%f", buyPrice)), nil)
	if err != nil {
		return err
	}

	err = db.Put([]byte("amount"), []byte(fmt.Sprintf("%f", amount)), nil)
	if err != nil {
		return err
	}

	err = db.Put([]byte("sellMarket"), []byte(sellMarket), nil)
	if err != nil {
		return err
	}

	err = db.Put([]byte("maxPrice"), []byte(fmt.Sprintf("%f", sellPrice)), nil)
	if err != nil {
		return err
	}

	return nil
}


func FindOpenOrders(db *leveldb.DB) (OrderData, bool, error){

    // Get data from the database
	output := OrderData{}
    res, err := db.Get([]byte("minPrice"), nil)
    if err != nil {
        return OrderData{}, true, err
    }
	min, err := strconv.ParseFloat(string(res), 64)
	if err != nil {
		return OrderData{}, true, err
	}
	if min == 0 {
		return OrderData{}, false, nil
	}
	output.BuyPrice = min

    res, err = db.Get([]byte("maxPrice"), nil)
    if err != nil {
        return OrderData{}, true, err
    }
	max, err := strconv.ParseFloat(string(res), 64)
	if err != nil {
		return OrderData{}, true, err
	}
	if max == 0 {
		return OrderData{}, false, nil
	}
	output.SellPrice = max

    res, err = db.Get([]byte("amount"), nil)
    if err != nil {
        return OrderData{}, true, err
    }
	amount, err := strconv.ParseFloat(string(res), 64)
	if err != nil {
		return OrderData{}, true, err
	}
	if amount == 0 {
		return OrderData{}, false, nil
	}
	output.Amount = amount

    res, err = db.Get([]byte("buyMarket"), nil)
    if err != nil {
        return OrderData{}, true, err
    }
	bmin := string(res)
	if bmin == "" {
		return OrderData{}, true, nil
	}

	output.BuyMarket = bmin

    res, err = db.Get([]byte("sellMarket"), nil)
    if err != nil {
        return OrderData{}, true, err
    }
	smin := string(res)
	if smin == "" {
		return OrderData{}, false, nil
	}

	output.SellMarket = smin

    res, err = db.Get([]byte("coin"), nil)
    if err != nil {
        return OrderData{}, true, err
    }
	coin := string(res)
	if coin == "" {
		return OrderData{}, false, nil
	}

	output.Coin = coin
    
	return output, true, nil
}

// removeLockFile removes the LOCK file if it exists.
func removeLockFile(dbPath string) {
	lockFilePath := dbPath + "/LOCK"
	if _, err := os.Stat(lockFilePath); err == nil {
		err = os.Remove(lockFilePath)
		if err != nil {
			log.Fatalf("Failed to remove LOCK file: %v", err)
		}
	}
}

// openLevelDBWithRetry attempts to open the LevelDB database with a retry mechanism.
func openLevelDBWithRetry(dbPath string, retries int) (*leveldb.DB, error) {
	var db *leveldb.DB
	var err error

	for i := 0; i < retries; i++ {
		db, err = leveldb.OpenFile(dbPath, nil)
		if err == nil {
			return db, nil
		}

		if errors.IsCorrupted(err) {
			log.Println("LevelDB is corrupted, trying recovery...")
			db, err = leveldb.RecoverFile(dbPath, nil)
			if err != nil {
				log.Fatalf("Failed to recover LevelDB: %v", err)
			}
			return db, nil
		}

		log.Printf("Failed to open LevelDB, attempt %d/%d: %v", i+1, retries, err)
		time.Sleep(2 * time.Second) // Wait before retrying
	}

	return nil, err
}
