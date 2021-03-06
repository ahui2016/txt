package main

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/ahui2016/txt/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	sessionName    = "txt-session"
	cookieSignIn   = "txt-cookie-signin"
	passwordMaxTry = 5
	allIP_MaxTry   = 100
	day            = 24 * 60 * 60
	defaultMaxAge  = 30 * day
)

var ipTryCount = make(map[string]int)

func checkIPTryCount(ip string) error {
	if *demo {
		return nil // 演示版允许无限重试密码
	}
	if ipTryCount[ip] >= passwordMaxTry || ipTryCount["all"] >= allIP_MaxTry {
		return fmt.Errorf("no more try, input wrong password too many times")
	}
	return nil
}

// checkPwdAndIP 检查 IP 与主密码，返回 true 表示有错误。
func checkPwdAndIP(c *gin.Context, pwd string) (exit bool) {
	ip := c.ClientIP()
	if err := checkIPTryCount(ip); err != nil {
		c.JSON(http.StatusForbidden, Text{err.Error()})
		return true
	}
	if pwd != db.Config.Password {
		ipTryCount[ip]++
		ipTryCount["all"]++
		c.JSON(http.StatusUnauthorized, Text{"wrong password"})
		return true
	}
	ipTryCount[ip] = 0
	return false
}

// checkKeyAndIP 检查 IP 与日常操作密钥，返回 true 表示有错误。
func checkKeyAndIP(c *gin.Context, secretKey string) (exit bool) {
	ip := c.ClientIP()
	if err := checkIPTryCount(ip); err != nil {
		c.JSON(http.StatusForbidden, Text{err.Error()})
		return true
	}
	if err := db.CheckKey(secretKey); err != nil {
		ipTryCount[ip]++
		ipTryCount["all"]++
		c.JSON(http.StatusUnauthorized, Text{err.Error()})
		return true
	}
	ipTryCount[ip] = 0
	return false
}

func CliCheckKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		var form SignInForm
		if BindCheck(c, &form) {
			c.Abort()
			return
		}
		if checkKeyAndIP(c, form.Password) {
			c.Abort()
			return
		}
		c.Next()
	}
}

func isSignedIn(c *gin.Context) bool {
	session := sessions.Default(c)
	yes, _ := session.Get(cookieSignIn).(bool)
	return yes
}

func CheckSignIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isSignedIn(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, Text{"require sign-in"})
			return
		}
		c.Next()
	}
}

func generateRandomKey() []byte {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	util.Panic(err)
	return b
}

func newNormalOptions() sessions.Options {
	return newOptions(defaultMaxAge)
}

func newExpireOptions() sessions.Options {
	return newOptions(-1)
}

func newOptions(maxAge int) sessions.Options {
	return sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func sessionSet(s sessions.Session, val bool, options sessions.Options) error {
	s.Set(cookieSignIn, val)
	s.Options(options)
	return s.Save()
}
