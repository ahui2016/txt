package mydb

import (
	"fmt"
	"time"

	"github.com/ahui2016/txt/model"
	"github.com/ahui2016/txt/util"
	bolt "go.etcd.io/bbolt"
)

type (
	TxtMsg = model.TxtMsg
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

func (db *DB) InsertTxtMsg(tm TxtMsg) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		return txPutObject(tx, temp_bucket, tm.ID, tm)
	})
}

func (db *DB) UpdateTxtMsg(tm TxtMsg) error {
	// 要注意 Alias 的新增/修改/删除 都要分别处理。
	return nil
}

// func (db *DB) getTxtMsgLimit(bucket string, limit int) (items []TxtMsg, err error) {
// 	i := 0
// 	err = db.DB.View(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte(bucket))
// 		return b.ForEach(func(k, v []byte) error {
// 			if i >= limit {
// 				return nil
// 			}
// 			tm, err := model.UnmarshalTxtMsg(v)
// 			if err != nil {
// 				return err
// 			}
// 			items = append(items, tm)
// 			i++
// 			return nil
// 		})
// 	})
// 	return
// }

func (db *DB) getTxtMsgLimit(bucket, start string, limit int) (items []TxtMsg, err error) {
	i := 0
	err = db.DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		for k, v := c.Seek([]byte(start)); k != nil; k, v = c.Next() {
			if i >= limit {
				break
			}
			if string(k) == start {
				continue
			}
			tm, err := model.UnmarshalTxtMsg(v)
			if err != nil {
				return err
			}
			items = append(items, tm)
			i++
		}
		return nil
	})
	return
}

func (db *DB) GetRecentItems() ([]TxtMsg, error) {
	limit := 15
	tempItems, err := db.getTxtMsgLimit(temp_bucket, "", limit)
	if err != nil {
		return nil, err
	}
	permItems, err := db.getTxtMsgLimit(perm_bucket, "", limit)
	if err != nil {
		return nil, err
	}
	return append(tempItems, permItems...), nil
}
