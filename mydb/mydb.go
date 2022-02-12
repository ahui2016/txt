package mydb

import (
	"database/sql"
	"fmt"

	"github.com/ahui2016/txt/util"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	Path string
	DB   *sql.DB
}

func (db *DB) Open(dbPath string) (err error) {
	db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1")
	e1 := initFirstID(txt_id_key, txt_id_prefix, db.DB)
	// e2 := db.initSettings(Settings{})
	return util.WrapErrors(e1)
}

func (db *DB) mustBegin() *sql.Tx {
	tx, err := db.DB.Begin()
	util.Panic(err)
	return tx
}

func (db *DB) Exec(query string, args ...interface{}) (err error) {
	_, err = db.DB.Exec(query, args...)
	return
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
