package main

import (
	"embed"
	"log"
	"net/http"

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

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/public/index.html")
	})

	auth := r.Group("/auth", Sleep())
	{
		auth.GET("/is-signed-in", func(c *gin.Context) {
			c.JSON(OK, isSignedIn(c))
		})
		auth.POST("/sign-in", signInHandler)
		auth.GET("/sign-out", signOutHandler)
		auth.POST("/get-current-key", getCurrentKey)
		auth.POST("/gen-new-key", generateKeyHandler)
		auth.POST("/change-pwd", changePwdHandler)
	}

	api := r.Group("/api", Sleep(), CheckSignIn())
	{
		api.POST("/add", addTxtMsg)
		api.GET("/recent-items", getRecentItems)
		api.POST("/toggle-category", toggleCatHandler)
		api.POST("/delete", deleteHandler)
		api.POST("/get-by-id", getByID)
		api.POST("/edit", editHandler)
		api.GET("/get-config", getConfig)
		api.POST("/update-config", updateConfig)
		api.POST("/get-more-items", getMoreItems)
		api.GET("/get-all-aliases", getAliasesHandler)
	}

	if err := r.Run(*addr); err != nil {
		log.Fatal(err)
	}
}
