package model

type TxtMsg struct {
	ID    string // ShortID
	Msg   string
	CTime int64
	MTime int64
}

type Settings struct {
	Password  string // 主密码，唯一作用是生成 Key
	Key       string // 日常使用的密钥
	KeyStarts int64  // Key 的生效时间 (timestamp)
	KeyMaxAge int64  // Key 的有效期（秒）
}
