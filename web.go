package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var web embed.FS

func static(s *gin.RouterGroup) {
	s.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=2592000")
	})

	s.GET("/robots.txt", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("User-agent: *\nAllow: /"))
	})

	s.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(200, "image/x-icon", []byte{})
	})

	assets, err := fs.Sub(web, "dist/assets")
	if err != nil {
		log.Fatalln("error:", err)
	}

	s.StaticFS("/assets/", http.FS(assets))
}

// todo 暂定 SPA，url 页面组装完整 html 返回，其余页面由 js 替换页面组件实现
func renderHtml(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {})
	r.GET("/av", func(c *gin.Context) {})
	r.GET("/cv", func(c *gin.Context) {})
	r.GET("/av/:aid", func(c *gin.Context) {})
	r.GET("/cv/:cid", func(c *gin.Context) {})
}
