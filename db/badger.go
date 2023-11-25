package db

import (
	"log"
	"os"
	"path"

	badger "github.com/dgraph-io/badger/v4"
)

var Badger Database

type Database struct {
	*badger.DB
}

func Init() {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	opts := badger.DefaultOptions(path.Join(homeDir, ".maestro", "database"))
	opts.Logger = nil // Disable logging
	Badger.DB, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
}

func (d *Database) Set(key, value string) error {
	return d.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

func (d *Database) Get(key string) (string, error) {
	var value string
	err := d.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			value = string(val)
			return nil
		})
		return err
	})
	return value, err
}
