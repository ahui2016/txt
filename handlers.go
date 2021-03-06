package main

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"time"

	"github.com/ahui2016/txt/model"
	"github.com/ahui2016/txt/mydb"
	"github.com/ahui2016/txt/util"
	"github.com/gin-contrib/sessions"
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

func signInHandler(c *gin.Context) {
	if isSignedIn(c) {
		c.Status(OK)
		return
	}
	var form SignInForm
	if BindCheck(c, &form) {
		return
	}
	if checkKeyAndIP(c, form.Password) {
		return
	}
	session := sessions.Default(c)
	checkErr(c, sessionSet(session, true, newNormalOptions()))
}

func signOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	checkErr(c, sessionSet(session, false, newExpireOptions()))
}

type secretKey struct {
	Key     string
	Starts  int64
	MaxAge  int64
	Expires int64
	IsGood  bool // 是否有效
}

func writeKeyResult(c *gin.Context, config model.Config) {
	expires := config.KeyStarts + config.KeyMaxAge
	c.JSON(OK, secretKey{
		Key:     config.Key,
		Starts:  config.KeyStarts,
		MaxAge:  config.KeyMaxAge / day, // 转换单位“天”
		Expires: expires,
		IsGood:  (util.TimeNow() <= expires),
	})
}

func getCurrentKey(c *gin.Context) {
	var form SignInForm
	if BindCheck(c, &form) {
		return
	}
	if checkPwdAndIP(c, form.Password) {
		return
	}
	writeKeyResult(c, db.Config)
}

func generateKeyHandler(c *gin.Context) {
	if *demo {
		c.JSON(500, Text{"Demo Mode (演示模式) 不可更新密钥。"})
		return
	}
	var form SignInForm
	if BindCheck(c, &form) {
		return
	}
	if checkPwdAndIP(c, form.Password) {
		return
	}
	if checkErr(c, db.GenNewKey()) {
		return
	}
	writeKeyResult(c, db.Config)
}

type ChangePwdForm struct {
	CurrentPwd string `form:"oldpwd" binding:"required"`
	NewPwd     string `form:"newpwd" binding:"required"`
}

func changePwdHandler(c *gin.Context) {
	if *demo {
		c.JSON(500, Text{"Demo Mode (演示模式) 不可更改主密码。"})
		return
	}
	var form ChangePwdForm
	if BindCheck(c, &form) {
		return
	}
	if checkPwdAndIP(c, form.CurrentPwd) {
		return
	}
	checkErr(c, db.ChangePwd(form.CurrentPwd, form.NewPwd))
}

func addTxtMsg(c *gin.Context) {
	type form struct {
		Msg string `form:"msg" binding:"required"`
	}
	var f form
	if BindCheck(c, &f) {
		return
	}
	msg, err := db.NewTxtMsg(f.Msg)
	if checkErr(c, err) {
		return
	}
	checkErr(c, db.InsertTxtMsg(msg))
}

func getRecentItems(c *gin.Context) {
	items, err := db.GetRecentItems(15)
	if checkErr(c, err) {
		return
	}
	c.JSON(OK, items)
}

func getMoreItems(c *gin.Context) {
	type form struct {
		Bucket string `form:"bucket" binding:"required"`
		Start  string `form:"start"`
		Limit  int    `form:"limit"`
	}
	var f form
	if BindCheck(c, &f) {
		return
	}
	items, err := db.GetMoreItems(f.Bucket, f.Start, f.Limit)
	if checkErr(c, err) {
		return
	}
	c.JSON(OK, items)
}

func cliGetMoreItems(c *gin.Context) {
	type form struct {
		Bucket string `form:"bucket" binding:"required"`
		Index  int    `form:"index"`
		Limit  int    `form:"limit" binding:"required"`
	}
	var f form
	if BindCheck(c, &f) {
		return
	}
	items, err := db.CliGetTxtMsg(f.Bucket, f.Index, f.Limit)
	if checkErr(c, err) {
		return
	}
	c.JSON(OK, items)
}

type AliasIndexForm struct {
	A_or_I string `form:"a_or_i" binding:"required"`
}

func getByAliasIndex(c *gin.Context) {
	var f AliasIndexForm
	if BindCheck(c, &f) {
		return
	}
	tm, err := db.GetByAliasIndex(f.A_or_I)
	if checkErr(c, err) {
		return
	}
	c.JSON(OK, tm)
}

func toggleCatHandler(c *gin.Context) {
	var f idForm
	if BindCheck(c, &f) {
		return
	}
	tm, err := db.GetByID(f.ID)
	if checkErr(c, err) {
		return
	}
	_, err = db.ToggleCat(tm)
	checkErr(c, err)
}

func cliToggleCat(c *gin.Context) {
	var f AliasIndexForm
	if BindCheck(c, &f) {
		return
	}
	tm, err := db.GetByAliasIndex(f.A_or_I)
	if checkErr(c, err) {
		return
	}
	after, err := db.ToggleCat(tm)
	if checkErr(c, err) {
		return
	}

	after.Index = 1
	c.JSON(OK, after)
}

func deleteHandler(c *gin.Context) {
	var f idForm
	if BindCheck(c, &f) {
		return
	}
	checkErr(c, db.DeleteTxtMsg(f.ID))
}

func cliDeleteHandler(c *gin.Context) {
	var f AliasIndexForm
	if BindCheck(c, &f) {
		return
	}
	checkErr(c, db.CliDeleteTxtMsg(f.A_or_I))
}

func getByID(c *gin.Context) {
	var f idForm
	if BindCheck(c, &f) {
		return
	}
	tm, err := db.GetByID(f.ID)
	if checkErr(c, err) {
		return
	}
	c.JSON(OK, tm)
}

func editHandler(c *gin.Context) {
	var f model.EditForm
	if BindCheck(c, &f) {
		return
	}
	err := db.Edit(f)
	if errors.Is(err, mydb.ErrKeyExists) {
		c.JSON(400, Text{"Alias Exists (别名冲突)"})
		return
	}
	checkErr(c, err)
}

func cliSetAlias(c *gin.Context) {
	type form struct {
		A_or_I string `form:"a_or_i" binding:"required"`
		Alias  string `form:"alias"`
	}
	var f form
	if BindCheck(c, &f) {
		return
	}
	checkErr(c, db.UpdateAlias(f.A_or_I, f.Alias))
}

func getConfig(c *gin.Context) {
	c.JSON(OK, db.Config.ToConfigForm())
}

func updateConfig(c *gin.Context) {
	if *demo {
		c.JSON(500, Text{"Demo Mode (演示模式) 不可修改配置"})
		return
	}
	var f model.ConfigForm
	if BindCheck(c, &f) {
		return
	}
	ignore, err := db.UpdateConfig(f)
	if checkErr(c, err) {
		return
	}
	c.JSON(OK, Text{ignore})
}

func getAliasesHandler(c *gin.Context) {
	aliases, err := db.GetAllAliases()
	if checkErr(c, err) {
		return
	}
	c.JSON(OK, aliases)
}

func searchHandler(c *gin.Context) {
	type form struct {
		Keyword string   `form:"keyword" binding:"required"`
		Buckets []string `form:"buckets"`
	}
	var f form
	if BindCheck(c, &f) {
		return
	}
	items, err := db.SearchTxtMsg(f.Keyword, f.Buckets)
	if checkErr(c, err) {
		return
	}
	c.JSON(OK, items)
}
