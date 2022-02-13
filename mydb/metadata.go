package mydb

import (
	"errors"

	"github.com/ahui2016/txt/model"
	"github.com/ahui2016/txt/util"

	"github.com/vmihailenco/msgpack/v5"
	bolt "go.etcd.io/bbolt"
)

// https://github.com/vmihailenco/msgpack

const (
	txtmsg_bucket    = "txtmsg-bucket"
	settings_bucket  = "settings-bucket"
	txt_id_key       = "txt-id-key"
	txt_id_prefix    = "T"
	settings_key     = "settings-key"
	hour             = 60 * 60
	day              = 24 * hour
	DefaultKeyMaxAge = 30 * day
)

var defaultSettings = Settings{
	Password:  "abc",
	Key:       util.RandomString(12), // 不需要太高的安全性
	KeyStarts: util.TimeNow(),
	KeyMaxAge: DefaultKeyMaxAge,
}

var ErrNoResult = errors.New("Error_DB_NoResult")

type (
	Settings = model.Settings
)

func txCreateBucket(tx *bolt.Tx, name string) error {
	_, err := tx.CreateBucketIfNotExists([]byte(name))
	return err
}

func (db *DB) CreateBuckets() error {
	tx := db.BeginWrite()
	defer tx.Rollback()

	e1 := txCreateBucket(tx, settings_bucket)
	e2 := txCreateBucket(tx, txtmsg_bucket)
	if err := util.WrapErrors(e1, e2); err != nil {
		return err
	}
	return tx.Commit()
}

func putBytes(bucket *bolt.Bucket, key string, v []byte) error {
	return bucket.Put([]byte(key), v)
}

func putString(bucket *bolt.Bucket, key string, v string) error {
	return bucket.Put([]byte(key), []byte(v))
}

func txPutBytes(tx *bolt.Tx, bucket, key string, v []byte) error {
	b := tx.Bucket([]byte(bucket))
	return b.Put([]byte(key), v)
}

func txPutString(tx *bolt.Tx, bucket, key string, v string) error {
	b := tx.Bucket([]byte(bucket))
	return b.Put([]byte(key), []byte(v))
}

func txPutObject(tx *bolt.Tx, bucket, key string, v interface{}) error {
	b := tx.Bucket([]byte(bucket))
	data, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), data)
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

func (db *DB) getString(bucket, key string) (string, error) {
	v, err := db.getBytes(bucket, key)
	return string(v), err
}

func (db *DB) GetSettings() (s Settings, err error) {
	data, err := db.getBytes(settings_bucket, settings_key)
	if err != nil {
		return
	}
	err = msgpack.Unmarshal(data, &s)
	return
}

func (db *DB) getCurrentID(idkey string) (id model.ShortID, err error) {
	strID, err := db.getString(settings_bucket, idkey)
	if err != nil {
		return
	}
	return model.ParseID(strID)
}

func (db *DB) initFirstID(idkey, prefix string) error {
	if _, err := db.getCurrentID(idkey); err != ErrNoResult {
		return err
	}
	id, err := model.FirstID(prefix)
	if err != nil {
		return err
	}
	return db.DB.Update(func(tx *bolt.Tx) error {
		return txPutString(tx, settings_bucket, txt_id_key, id.String())
	})
}

// genNextID generates a new ShortID.
func (db *DB) genNextID(idkey string) (nextID string, err error) {
	currentID, err := db.getCurrentID(idkey)
	if err != nil {
		return
	}
	nextID = currentID.Next().String()
	err = db.DB.Update(func(tx *bolt.Tx) error {
		return txPutString(tx, settings_bucket, idkey, nextID)
	})
	return
}

func (db *DB) initSettings() error {
	if _, err := db.GetSettings(); err != ErrNoResult {
		return err
	}
	return db.DB.Update(func(tx *bolt.Tx) error {
		return txPutObject(tx, settings_bucket, settings_key, defaultSettings)
	})
}

func (db *DB) UpdateSettings(s Settings) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		return txPutObject(tx, settings_bucket, settings_key, s)
	})
}
