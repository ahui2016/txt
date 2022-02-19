package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// WrapErrors 把多个错误合并为一个错误.
func WrapErrors(allErrors ...error) (wrapped error) {
	for _, err := range allErrors {
		if err != nil {
			if wrapped == nil {
				wrapped = err
			} else {
				wrapped = fmt.Errorf("%s | %w", wrapped, err)
			}
		}
	}
	return
}

// Panic panics if err != nil
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

func PathIsNotExist(name string) (ok bool) {
	_, err := os.Lstat(name)
	if os.IsNotExist(err) {
		ok = true
		err = nil
	}
	Panic(err)
	return
}

func PathIsExist(name string) bool {
	return !PathIsNotExist(name)
}

func TimeNow() int64 {
	return time.Now().Unix()
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// Marshal64 converts data to json and encodes to base64 string.
func Marshal64(data interface{}) (string, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return Base64Encode(dataJSON), err
}

// Unmarshal64_Wrong 是一个错误的的函数，不可使用！
// 因为 value 是值，不是指针，因此 &value 无法传出去。
func Unmarshal64_Wrong(data64 string, value interface{}) error {
	data, err := Base64Decode(data64)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &value)
}

func RandomBytes(size int) []byte {
	someBytes := make([]byte, size)
	_, err := rand.Read(someBytes)
	Panic(err)
	return someBytes
}

// RandomString 的 size 不是最终长度
func RandomString(size int) string {
	someBytes := RandomBytes(size)
	return Base64Encode(someBytes)
}
