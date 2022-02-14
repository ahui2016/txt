package mydb

import (
	"fmt"
	"time"

	"github.com/ahui2016/txt/util"
	bolt "go.etcd.io/bbolt"
)

type DB struct {
	Path   string
	DB     *bolt.DB
	Config Config
}

func (db *DB) Open(dbPath string) (err error) {
	db.DB, err = bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	db.Path = dbPath
	e1 := db.createBuckets()
	e2 := db.initConfig()
	return util.WrapErrors(e1, e2)
}

func (db *DB) BeginWrite() *bolt.Tx {
	tx, err := db.DB.Begin(true)
	util.Panic(err)
	return tx
}

func (db *DB) BeginWriteBucket(name string) (*bolt.Tx, *bolt.Bucket) {
	tx, err := db.DB.Begin(true)
	util.Panic(err)
	b := tx.Bucket([]byte(name))
	return tx, b
}

func (db *DB) BeginRead() *bolt.Tx {
	tx, err := db.DB.Begin(false)
	util.Panic(err)
	return tx
}

func (db *DB) CheckKey(key string) error {
	if key != db.Config.Key {
		return fmt.Errorf("wrong key")
	}
	if util.TimeNow() > db.Config.KeyStarts+db.Config.KeyMaxAge {
		return fmt.Errorf("the key is expired")
	}
	return nil
}
