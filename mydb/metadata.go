package mydb

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

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
var ErrKeyExists = errors.New("error-database-key-exists")
var ErrMsgTooLong = errors.New("error-message-too-long")

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

func getBucketName(tm TxtMsg) string {
	if tm.Cat == CatTemp {
		return temp_bucket
	}
	return perm_bucket
}

func bucketPutObject(bucket *bolt.Bucket, key string, v interface{}) error {
	data, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return bucket.Put([]byte(key), data)
}

func txPutObject(tx *bolt.Tx, bucket, key string, v interface{}) error {
	b := tx.Bucket([]byte(bucket))
	return bucketPutObject(b, key, v)
}

// txLimitTemp 限制 temp_bucket 中的数量，如果达到 limit 就删除旧条目。
// 即, txLimitTemp 执行后，temp_bucket 中的条目数量应小于 limit (而不是小于等于 limit)。
// 通常在 bucket.Put 之前执行本函数，即, bucket.Put 之后的条目数量小于等于 limit。
func txLimitTemp(tx *bolt.Tx, limit int) error {
	if limit < 1 {
		return nil
	}
	bucket := tx.Bucket([]byte(temp_bucket))
	c := bucket.Cursor()

	// 特殊情况优化 1. 如果 temp_bucket 中的条目数量小于 limit，则不需要删除任何条目。
	n := bucket.Stats().KeyN
	log.Print("temp_bucket.Stats().KeyN: ", n)
	log.Print("limit: ", limit)
	if n < limit {
		return nil
	}

	// 特殊情况优化 2. 如果 temp_bucket 中的条目数量刚好等于 limit，
	// 则只要删除最早的 1 个条目。
	if n == limit {
		k, _ := c.First()
		log.Print("First key: ", k)
		return bucket.Delete(k)
	}

	// 普通情况（无法优化的情况）
	i := 1
	for k, _ := c.Last(); k != nil; k, _ = c.Prev() {
		log.Printf("i: %d, limit: %d", i, limit)
		if i < limit {
			i++
			continue
		}
		if err := bucket.Delete(k); err != nil {
			log.Print("Delete key: ", k)
			return err
		}
	}
	return nil
}

func txGetBytes(tx *bolt.Tx, bucket, key string) ([]byte, error) {
	b := tx.Bucket([]byte(bucket))
	v := b.Get([]byte(key))
	if v == nil {
		return nil, ErrNoResult
	}
	return v, nil
}

func (db *DB) getBytes(bucket, key string) (v []byte, err error) {
	err = db.DB.View(func(tx *bolt.Tx) error {
		v, err = txGetBytes(tx, bucket, key)
		return err
	})
	return
}

func txGetByID(tx *bolt.Tx, id string) (tm TxtMsg, err error) {
	data, err := txGetBytes(tx, temp_bucket, id)
	if err != nil && err != ErrNoResult {
		return
	}
	if err == ErrNoResult {
		if data, err = txGetBytes(tx, perm_bucket, id); err != nil {
			return
		}
	}
	// 此时 err == nil, 并且 data 也获得了内容。
	err = msgpack.Unmarshal(data, &tm)
	return
}

// func (db *DB) bucketGetString(bucket *bolt.Bucket, key string) (string, error) {
// 	v, err := db.getBytes(bucket, key)
// 	return string(v), err
// }

// func bucketGetTxtMsg(bucket *bolt.Bucket, key []byte) (tm TxtMsg, err error) {
// 	data := bucket.Get([]byte(key))
// 	if data == nil {
// 		return tm, ErrNoResult
// 	}
// 	err = msgpack.Unmarshal(data, &tm)
// 	return
// }

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

// UpdateConfig updates the config from a ConfigForm.
func (db *DB) UpdateConfig(cf model.ConfigForm) (warning string, err error) {
	var ignore []string
	config := db.Config

	maxAge := cf.KeyMaxAge * day
	if maxAge < 1 {
		ignore = append(ignore, "key_max_age")
	} else {
		config.KeyMaxAge = cf.KeyMaxAge * day
	}

	if cf.MsgSizeLimit < 256 {
		ignore = append(ignore, "msg_size_limit")
	} else {
		config.MsgSizeLimit = cf.MsgSizeLimit
	}

	if cf.TempLimit < 1 {
		ignore = append(ignore, "temp_msg_limit")
	} else {
		config.TempLimit = cf.TempLimit
	}

	if cf.EveryPageLimit < 1 {
		ignore = append(ignore, "page_limit")
	} else {
		config.EveryPageLimit = cf.EveryPageLimit
	}

	if _, err = model.DateID(cf.TimeOffset); err != nil {
		ignore = append(ignore, "timeone_offset")
	} else {
		config.TimeOffset = cf.TimeOffset
	}

	if err = db.updateConfig(config); err != nil {
		return
	}
	if len(ignore) > 0 {
		warning = "ignore: " + strings.Join(ignore, ", ")
	}
	// 要记得更新 db.Config
	db.Config = config
	return
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

// ChangePassword 修改密码，其中 oldPwd 由于涉及 ip 尝试次数，因此应在
// 使用本函数前使用 db.CheckPassword 验证 oldPwd.
func (db *DB) ChangePwd(oldPwd, newPwd string) error {
	if oldPwd == "" {
		return fmt.Errorf("the current password is empty")
	}
	if newPwd == "" {
		return fmt.Errorf("the new password is empty")
	}
	if newPwd == oldPwd {
		return fmt.Errorf("the two passwords are the same")
	}

	config, err := db.getConfig()
	if err != nil {
		return err
	}
	if config.Password != oldPwd {
		return fmt.Errorf("the current password is wrong")
	}
	config.Password = newPwd
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

func (db *DB) newDateID() (string, error) {
	return model.DateID(db.Config.TimeOffset)
}

func (db *DB) NewTxtMsg(msg string) (TxtMsg, error) {
	if len(msg) > db.Config.MsgSizeLimit {
		err := fmt.Errorf("size: %d, limit: %d", len(msg), db.Config.MsgSizeLimit)
		return TxtMsg{}, util.WrapErrors(ErrMsgTooLong, err)
	}
	return model.NewTxtMsg(msg, db.Config.TimeOffset)
}

func bucketUpdateIndex(bucket *bolt.Bucket) error {
	max := bucket.Stats().KeyN
	err := bucket.ForEach(func(k, v []byte) error {
		tm, err := model.UnmarshalTxtMsg(v)
		if err != nil {
			return err
		}
		tm.Index = max
		max--
		return bucketPutObject(bucket, string(k), tm)
	})
	return err
}

// 注意：由于 bucketUpdateIndex 涉及 bucket.Stats().KeyN,
// 必须在前面的删除/插入 commit 之后, bucket.Stats() 才会更新。
func (db *DB) updateIndex(bucket string) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return bucketUpdateIndex(b)
	})
	return err
}

// 注意：由于 bucketUpdateIndex 涉及 bucket.Stats().KeyN,
// 必须在前面的删除/插入 commit 之后, bucket.Stats() 才会更新。
func (db *DB) updateAllIndex() error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		b1 := tx.Bucket([]byte(temp_bucket))
		if err := bucketUpdateIndex(b1); err != nil {
			return err
		}
		b2 := tx.Bucket([]byte(perm_bucket))
		return bucketUpdateIndex(b2)
	})
	return err
}

func txPutAlias(tx *bolt.Tx, alias, id string, overwrite bool) error {
	b := tx.Bucket([]byte(alias_bucket))
	if !overwrite && b.Get([]byte(alias)) != nil {
		return ErrKeyExists
	}
	return b.Put([]byte(alias), []byte(id))
}
func txDeleteAlias(tx *bolt.Tx, alias string) error {
	b := tx.Bucket([]byte(alias_bucket))
	return b.Delete([]byte(alias))
}
func txChangeAlias(tx *bolt.Tx, oldAlias, newAlias string) error {
	// 确保新旧别名都不是空字符串
	if oldAlias == "" || newAlias == "" {
		return fmt.Errorf("old alias or new alias is empty")
	}
	// 确保旧别名存在
	b := tx.Bucket([]byte(alias_bucket))
	id := b.Get([]byte(oldAlias))
	if id == nil {
		return ErrNoResult
	}
	// 确保新别名无冲突
	if b.Get([]byte(newAlias)) != nil {
		return ErrKeyExists
	}
	// 插入新别名
	if err := b.Put([]byte(newAlias), id); err != nil {
		return err
	}
	// 删除旧别名
	return b.Delete([]byte(oldAlias))
}

// txEditAlias 专用于 DB.Edit(model.EditForm)
func txEditAlias(tx *bolt.Tx, oldAlias, newAlias, id string) error {
	// 别名不可采用“以 T 或 P 开头紧跟数字”的形式（要避免与 index 冲突）
	if err := checkAlias(newAlias); err != nil {
		return err
	}
	// 从有别名变成无别名（即，删除别名）
	if oldAlias != "" && newAlias == "" {
		return txDeleteAlias(tx, oldAlias)
	}
	// 从无别名变成有别名（即，新增别名）
	if oldAlias == "" && newAlias != "" {
		return txPutAlias(tx, newAlias, id, false)
	}
	// 有别名但新旧别名不相同（即，更改别名）
	if oldAlias != newAlias {
		return txChangeAlias(tx, oldAlias, newAlias)
	}
	// 新旧别名相同
	return nil
}

func checkAlias(alias string) error {
	if alias == "" {
		return nil
	}
	alias = strings.ToUpper(alias)
	if alias[0] != 'T' && alias[0] != 'P' {
		return nil
	}
	alias = alias[1:]
	if _, err := strconv.Atoi(alias); err == nil {
		return fmt.Errorf("别名不可采用“以 T 或 P 开头紧跟数字”的形式")
	}
	return nil
}
