package mydb

import (
	"fmt"
	"time"

	"github.com/ahui2016/txt/util"
	_ "github.com/mattn/go-sqlite3"
	bolt "go.etcd.io/bbolt"
)

//TODO https://github.com/etcd-io/bbolt

type DB struct {
	Path string
	DB   *bolt.DB
}

func (db *DB) Open(dbPath string) (err error) {
	db.DB, err = bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	db.Path = dbPath
	e1 := db.initFirstID(txt_id_key, txt_id_prefix)
	e2 := db.initSettings()
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
	s, err := db.GetSettings()
	if err != nil {
		return err
	}
	if key != s.Key {
		return fmt.Errorf("wrong key")
	}
	if util.TimeNow() > s.KeyStarts+s.KeyMaxAge {
		return fmt.Errorf("the key is expired")
	}
	return nil
}
