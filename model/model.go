package model

import "time"

type Category string

const (
	CatTemp Category = "Category-Temporary"
	CatPerm Category = "Category-Permanent"
)

type TxtMsg struct {
	ID     string // DateID
	UserID string // 暂时不使用，以后升级为多用户系统时使用
	Alias  string // 别名，用于方便检索
	Msg    string
	Cat    Category // 类型（比如暂存、永久）
	CTime  int64
	MTime  int64
}

type Config struct {
	Password       string // 主密码，唯一作用是生成 Key
	Key            string // 日常使用的密钥
	KeyStarts      int64  // Key 的生效时间 (timestamp)
	KeyMaxAge      int64  // Key 的有效期（秒）
	MsgSizeLimit   int64  // 每条消息的长度上限
	TempLimit      int64  // 暂存消息条数上限（永久消息不设上限）
	EveryPageLimit int64  // 每页最多列出多少条消息

}

// DateID 返回一个便于通过前缀筛选时间范围的字符串 id,
// 由于精确到秒，为了避免重复，每次生成新 id 前会先暂停一秒。
func DateID() string {
	time.Sleep(time.Second)
	return time.Now().Format("20060102150405")
}
