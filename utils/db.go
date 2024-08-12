package utils

import (
    "log"
    "os"
    "time"

    "github.com/syndtr/goleveldb/leveldb"
    "github.com/syndtr/goleveldb/leveldb/errors"
)

func Database(dbPath string) (*leveldb.DB, error){
	removeLockFile(dbPath)

	// Attempt to open the database
	db, err := openLevelDBWithRetry(dbPath, 3)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	log.Println("LevelDB opened successfully")
	return db, nil
}

// removeLockFile removes the LOCK file if it exists.
func removeLockFile(dbPath string) {
	lockFilePath := dbPath + "/LOCK"
	if _, err := os.Stat(lockFilePath); err == nil {
		log.Println("LOCK file exists. Removing it...")
		err = os.Remove(lockFilePath)
		if err != nil {
			log.Fatalf("Failed to remove LOCK file: %v", err)
		}
		log.Println("LOCK file removed successfully")
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
