package model

import (
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Category string

const (
	CatTemp Category = "Category-Temporary"
	CatPerm Category = "Category-Permanent"
)

type TxtMsg struct {
	ID     string   // DateID, 既是 id 也是创建日期
	UserID string   // 暂时不使用，以后升级为多用户系统时使用
	Alias  string   // 别名，要注意与 Alias bucket 联动。
	Msg    string   // 消息内容
	Cat    Category // 类型（比如暂存、永久）
	Index  int      // 流水号，每当插入或删除条目时，需要更新全部条目的流水号
}

func NewTxtMsg(msg, offset string) (tm TxtMsg, err error) {
	id, err := DateID(offset)
	if err != nil {
		return
	}
	tm = TxtMsg{
		ID:  id,
		Msg: msg,
		Cat: CatTemp,
	}
	return
}

func UnmarshalTxtMsg(data []byte) (tm TxtMsg, err error) {
	err = msgpack.Unmarshal(data, &tm)
	return
}

// Alias 指向一条 TxtMsg, 要注意与 TxtMsg 联动（同时添加/修改/删除）。
type Alias struct {
	ID    string // TxtMsg.Alias
	MsgID string
}

type Config struct {
	Password       string // 主密码，唯一作用是生成 Key
	Key            string // 日常使用的密钥
	KeyStarts      int64  // Key 的生效时间 (timestamp)
	KeyMaxAge      int64  // Key 的有效期（秒）
	MsgSizeLimit   int64  // 每条消息的长度上限
	TempLimit      int64  // 暂存消息条数上限（永久消息不设上限）
	EveryPageLimit int64  // 每页最多列出多少条消息
	TimeOffset     string // "+8" 表示北京时间, "-5" 表示纽约时间, 依此类推。
}

// DateID 返回一个便于通过前缀筛选时间范围的字符串 id,
// 由于精确到秒，为了避免重复，每次生成新 id 前会先暂停一秒。
// offset 的格式是 "+8" 表示东八区(北京时间), "-5" 表示西五区(纽约时间), 依此类推。
// 返回的 id 格式是 "2006-01-02_150405", 由于有可能用作 html 元素的 id, 因此不含空格与冒号。
func DateID(offset string) (string, error) {
	time.Sleep(time.Second)

	timezone, err := time.ParseDuration(offset + "h")
	if err != nil {
		return "", err
	}

	// utcFormat 大概长这个样子 => "2022-02-14_214208+00:00"
	// 由于 utcFormat 包含了时区 +00:00, 因此 dt 的时区就是 UTC
	utcFormat := time.Now().UTC().Format("2006-01-02_150405-07:00")
	dt, err := time.Parse("2006-01-02_150405-07:00", utcFormat)
	if err != nil {
		return "", err
	}
	// 由于 dt 的时区是 UTC, 格式化是就是按照 UTC 来输出字符串的，
	// 因此可以通过加减时间来假装时区变更。
	return dt.Add(timezone).Format("2006-01-02_150405"), nil
}
