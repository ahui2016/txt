package main

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/ahui2016/txt/util"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

const OK = http.StatusOK

// Text 用于向前端返回一个简单的文本消息。
// 为了保持一致性，总是向前端返回 JSON, 因此即使是简单的文本消息也使用 JSON.
type Text struct {
	Message string `json:"message"`
}

func checkErr(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(500, Text{err.Error()})
		return true
	}
	return false
}

// BindCheck binds an obj, returns true if err != nil.
func BindCheck(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBind(obj); err != nil {
		c.JSON(400, Text{err.Error()})
		return true
	}
	return false
}

type Number struct {
	N int64 `json:"n"`
}

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

// https://github.com/gin-contrib/static/issues/19
// https://blog.carlmjohnson.net/post/2021/how-to-use-go-embed/
func EmbedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	fsys, err := fs.Sub(fsEmbed, targetPath)
	util.Panic(err)
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}

// Sleep 在 debug 模式中暂停一秒模拟网络缓慢的情形。
func Sleep() gin.HandlerFunc {
	return func(c *gin.Context) {
		if *debug {
			time.Sleep(time.Second)
		}
		c.Next()
	}
}

// JavaScriptHeader 确保向前端返回正确的 js 文件类型。
func JavaScriptHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/javascript")
		c.Next()
	}
}

type idForm struct {
	ID string `form:"id" binding:"required"`
}

type SignInForm struct {
	Password string `form:"password" binding:"required"`
}
