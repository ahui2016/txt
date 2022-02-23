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

var (
	ErrSameAsLast = fmt.Errorf("same as last message")
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

// InsertTxtMsg 注意此时必然插入到 temp_bucket, 并且 Alias 必然为空。
// 要注意暂存消息的数量上限。
func (db *DB) InsertTxtMsg(tm TxtMsg) error {
	if err := db.DB.Update(func(tx *bolt.Tx) error {
		last, err := db.getLastTempMsg()
		if err != nil {
			return err
		}
		// 如果新消息的内容刚好与最新一条暂存消息相同，则不插入。
		// 此时只管暂存消息，不管永久消息。
		if last.Msg == tm.Msg {
			return ErrSameAsLast
		}
		if err := txLimitTemp(tx, db.Config.TempLimit); err != nil {
			return err
		}
		return txPutObject(tx, temp_bucket, tm.ID, tm)
	}); err != nil {
		return err
	}
	return db.updateIndex(temp_bucket)
}

// DeleteTxtMsg 删除 id. 注意：如有 Alias 要同步删除。
func (db *DB) DeleteTxtMsg(id string) error {
	tm, err := db.GetByID(id)
	if err != nil {
		return err
	}
	err = db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(getBucketName(tm)))
		if err := b.Delete([]byte(id)); err != nil {
			return err
		}
		if tm.Alias != "" {
			if err := txDeleteAlias(tx, tm.Alias); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (db *DB) GetByID(id string) (tm TxtMsg, err error) {
	err = db.DB.View(func(tx *bolt.Tx) error {
		tm, err = txGetByID(tx, id)
		return err
	})
	return
}

func (db *DB) getLastTempMsg() (tm TxtMsg, err error) {
	err = db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(temp_bucket))
		k, v := b.Cursor().Last()
		if k == nil {
			return nil
		}
		return msgpack.Unmarshal(v, &tm)
	})
	return
}

// ToggleCat 在暂存消息与永久消息之间转换，为了让转换后的消息排在前面，
// 转换时会改变 ID, 又由于 ID 同时也是创建日期，因此相当于同时改变创建日期。
// 注意：如有 Alias 要同步更新 ID.
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
		if err := bucketPutObject(targetBucket, tm.ID, tm); err != nil {
			return err
		}
		if err := srcBucket.Delete(srcID); err != nil {
			return err
		}
		if tm.Alias != "" {
			err = txPutAlias(tx, tm.Alias, targetID, true)
		}
		return err
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
		k, v := c.Last()
		if start != "" {
			_, _ = c.Seek([]byte(start))
			k, v = c.Prev()
		}
		for ; k != nil; k, v = c.Prev() {
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

func (db *DB) getAliasLimit(start string, limit int) (items []TxtMsg, err error) {
	i := 0
	err = db.DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(alias_bucket)).Cursor()
		alias, id := c.First()
		if start != "" {
			_, _ = c.Seek([]byte(start))
			alias, id = c.Next()
		}
		for ; alias != nil; alias, id = c.Next() {
			if i >= limit {
				break
			}
			tm, err := txGetByID(tx, string(id))
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

func (db *DB) GetMoreItems(bucket, start string, limit int) ([]TxtMsg, error) {
	if limit <= 0 {
		limit = db.Config.EveryPageLimit
	}
	if bucket == alias_bucket {
		return db.getAliasLimit(start, limit)
	}
	return db.getTxtMsgLimit(bucket, start, limit)
}

// Edit from EditForm, 要注意同步更新 Alias.
func (db *DB) Edit(form model.EditForm) error {
	tm, err := db.GetByID(form.ID)
	if err != nil {
		return err
	}
	err = db.DB.Update(func(tx *bolt.Tx) error {
		if err := txEditAlias(tx, tm.Alias, form.Alias, form.ID); err != nil {
			return err
		}
		tm.Alias = form.Alias
		tm.Msg = form.Msg
		return txPutObject(tx, getBucketName(tm), tm.ID, tm)
	})
	return err
}

func (db *DB) GetAllAliases() (aliases []model.Alias, err error) {
	err = db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(alias_bucket))
		return b.ForEach(func(k, v []byte) error {
			aliases = append(aliases, model.Alias{
				ID:    string(k),
				MsgID: string(v),
			})
			return nil
		})
	})
	return aliases, err
}
