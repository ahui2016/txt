package mydb

import (
	"fmt"
	"time"

	"github.com/ahui2016/txt/model"
	"github.com/ahui2016/txt/util"
	"github.com/vmihailenco/msgpack/v5"
	bolt "go.etcd.io/bbolt"
)

type (
	TxtMsg = model.TxtMsg
)

const (
	CatTemp = model.CatTemp
	CatPerm = model.CatPerm
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
	if err := db.DB.Update(func(tx *bolt.Tx) error {
		return txPutObject(tx, temp_bucket, tm.ID, tm)
	}); err != nil {
		return err
	}
	return db.updateIndex(temp_bucket)
}

func (db *DB) UpdateTxtMsg(tm TxtMsg) error {
	// 要注意 Alias 的新增/修改/删除 都要分别处理。
	return nil
}

// DeleteTxtMsg 删除 id, 如果 id 不存在也返回 nil.
func (db *DB) DeleteTxtMsg(id string) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		b1 := tx.Bucket([]byte(temp_bucket))
		if err := b1.Delete([]byte(id)); err != nil {
			return err
		}
		b2 := tx.Bucket([]byte(perm_bucket))
		return b2.Delete([]byte(id))
	})
	return err
}

func (db *DB) GetByID(id string) (tm TxtMsg, err error) {
	data, err := db.getBytes(temp_bucket, id)
	if err != nil && err != ErrNoResult {
		return
	}
	if err == ErrNoResult {
		if data, err = db.getBytes(perm_bucket, id); err != nil {
			return
		}
	}
	// 此时 err == nil, 并且 data 也获得了内容。
	err = msgpack.Unmarshal(data, &tm)
	return
}

// ToggleCat 在暂存消息与永久消息之间转换，为了让转换后的消息排在前面，
// 转换时会改变 ID, 又由于 ID 同时也是创建日期，因此相当于同时改变创建日期。
func (db *DB) ToggleCat(tm TxtMsg) error {
	var srcBucket, targetBucket *bolt.Bucket
	var targetCat model.Category
	if err := db.DB.Update(func(tx *bolt.Tx) error {
		if tm.Cat == CatTemp {
			srcBucket = tx.Bucket([]byte(temp_bucket))
			targetBucket = tx.Bucket([]byte(perm_bucket))
			targetCat = CatPerm
		} else {
			srcBucket = tx.Bucket([]byte(perm_bucket))
			targetBucket = tx.Bucket([]byte(temp_bucket))
			targetCat = CatTemp
		}
		srcID := []byte(tm.ID)
		targetID, err := db.newDateID()
		if err != nil {
			return err
		}
		tm.ID = targetID
		tm.Cat = targetCat
		if err := bucketPutObject(targetBucket, []byte(tm.ID), tm); err != nil {
			return err
		}
		return srcBucket.Delete(srcID)
	}); err != nil {
		return err
	}
	// 注意：由于 db.updateAllIndex 涉及 bucket.Stats().KeyN,
	// 上面的删除/插入必须 commit 之后, bucket.Stats() 才会更新。
	return db.updateAllIndex()
}

func (db *DB) getTxtMsgLimit(bucket, start string, limit int) (items []TxtMsg, err error) {
	i := 0
	err = db.DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if i >= limit {
				break
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
