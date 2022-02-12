package main

import (
	"embed"
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

//go:embed static
var staticFiles embed.FS

//go:embed ts/dist/*.js
var staticJS embed.FS

func main() {
	defer db.DB.Close()

	if *debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
		log.Print("[Listen and serve] ", *addr)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	if *debug {
		r.Use(gin.Logger())
	}

	// 必须正确设置此项才能获取真实IP
	r.SetTrustedProxies([]string{"127.0.0.1"})

	sessionStore := cookie.NewStore(generateRandomKey())
	r.Use(sessions.Sessions(sessionName, sessionStore))

	// release mode 使用 embed 的文件，否则使用当前目录的 static 文件。
	if gin.Mode() == gin.ReleaseMode {
		r.StaticFS("/public", EmbedFolder(staticFiles, "static"))
		// 这个 Group 只是为了给 StaticFS 添加 middleware
		r.Group("/js", JavaScriptHeader()).
			StaticFS("/", EmbedFolder(staticJS, "ts/dist"))
	} else {
		r.Static("/public", "static")
		r.Group("/js", JavaScriptHeader()).Static("/", "ts/dist")
	}

	r.GET("/robots.txt", func(c *gin.Context) {
		c.FileFromFS("/robots.txt", EmbedFolder(staticFiles, "static"))
	})

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", EmbedFolder(staticFiles, "static"))
	})
}
