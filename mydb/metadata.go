package mydb

import (
	"errors"

	"github.com/ahui2016/txt/model"
	"github.com/ahui2016/txt/util"

	"github.com/vmihailenco/msgpack/v5"
	bolt "go.etcd.io/bbolt"
)

// https://github.com/etcd-io/bbolt
// https://github.com/vmihailenco/msgpack

const (
	config_key          = "config-key"
	temp_bucket         = "temporary-bucket"
	perm_bucket         = "permanent-bucket"
	alias_bucket        = "alias-bucket"
	config_bucket       = "config-bucket"
	txt_id_key          = "txt-id-key"
	txt_id_prefix       = "T"
	hour                = 60 * 60
	day                 = 24 * hour
	defaultKeyMaxAge    = 30 * day
	defaultMsgSizeLimit = 1024
	defaultTempLimit    = 100
	defaultPageLimit    = 30
	BeijingTime         = "+8" // 北京时间
	secretKeySize       = 12   // 不需要太高的安全性
)

var defaultConfig = Config{
	Password:       "abc",
	Key:            util.RandomString(secretKeySize),
	KeyStarts:      util.TimeNow(),
	KeyMaxAge:      defaultKeyMaxAge,
	MsgSizeLimit:   defaultMsgSizeLimit,
	TempLimit:      defaultTempLimit,
	EveryPageLimit: defaultPageLimit,
	TimeOffset:     BeijingTime,
}

var ErrNoResult = errors.New("error-database-no-result")

type (
	Config = model.Config
)

func txCreateBucket(tx *bolt.Tx, name string) error {
	_, err := tx.CreateBucketIfNotExists([]byte(name))
	return err
}

func (db *DB) createBuckets() error {
	tx := db.BeginWrite()
	defer tx.Rollback()

	e1 := txCreateBucket(tx, config_bucket)
	e2 := txCreateBucket(tx, temp_bucket)
	e3 := txCreateBucket(tx, perm_bucket)
	e4 := txCreateBucket(tx, alias_bucket)
	if err := util.WrapErrors(e1, e2, e3, e4); err != nil {
		return err
	}
	return tx.Commit()
}

// func putBytes(bucket *bolt.Bucket, key string, v []byte) error {
// 	return bucket.Put([]byte(key), v)
// }
// func putString(bucket *bolt.Bucket, key string, v string) error {
// 	return bucket.Put([]byte(key), []byte(v))
// }

// func txPutBytes(tx *bolt.Tx, bucket, key string, v []byte) error {
// 	b := tx.Bucket([]byte(bucket))
// 	return b.Put([]byte(key), v)
// }

// func (db *DB) getString(bucket, key string) (string, error) {
// 	v, err := db.getBytes(bucket, key)
// 	return string(v), err
// }

// func (db *DB) getInt64(bucket, key string) (n int64, err error) {
// 	data, err := db.getBytes(bucket, key)
// 	if err != nil {
// 		return
// 	}
// 	err = msgpack.Unmarshal(data, &n)
// 	return
// }

// func txPutString(tx *bolt.Tx, bucket, key string, v string) error {
// 	b := tx.Bucket([]byte(bucket))
// 	return b.Put([]byte(key), []byte(v))
// }

func bucketPutObject(bucket *bolt.Bucket, key []byte, v interface{}) error {
	data, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return bucket.Put(key, data)
}

func txPutObject(tx *bolt.Tx, bucket, key string, v interface{}) error {
	b := tx.Bucket([]byte(bucket))
	return bucketPutObject(b, []byte(key), v)
}

func (db *DB) getBytes(bucket, key string) (v []byte, err error) {
	err = db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v = b.Get([]byte(key))
		if v == nil {
			return ErrNoResult
		}
		return nil
	})
	return
}

func (db *DB) getConfig() (config Config, err error) {
	data, err := db.getBytes(config_bucket, config_key)
	if err != nil {
		return
	}
	err = msgpack.Unmarshal(data, &config)
	return
}

func (db *DB) initConfig() error {
	config, err := db.getConfig()
	if err == nil {
		db.Config = config
		return nil
	}
	if err != ErrNoResult {
		return err
	}
	// 剩下的唯一可能性就是 err == ErrNoResult
	return db.updateConfig(defaultConfig)
}

func (db *DB) updateConfig(config Config) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		return txPutObject(tx, config_bucket, config_key, config)
	})
	if err != nil {
		return err
	}
	// 要记得更新 db.Config
	db.Config = config
	return nil
}

func (db *DB) GenNewKey() error {
	config, err := db.getConfig()
	if err != nil {
		return err
	}
	config.Key = util.RandomString(secretKeySize)
	config.KeyStarts = util.TimeNow()
	return db.updateConfig(config)
}

func (db *DB) Count(bucket string) (n int) {
	_ = db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		n = b.Stats().KeyN
		return nil
	})
	return
}

func (db *DB) NewTxtMsg(msg string) (TxtMsg, error) {
	return model.NewTxtMsg(msg, db.Config.TimeOffset)
}

func (db *DB) updateIndex(bucket string) error {
	i := 0
	err := db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.ForEach(func(k, v []byte) error {
			i++
			tm, err := model.UnmarshalTxtMsg(v)
			if err != nil {
				return err
			}
			tm.Index = i
			return bucketPutObject(b, k, tm)
		})
		return err
	})
	return err
}
