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

	dist, err := fs.Sub(web, "dist")
	if err != nil {
		log.Fatalln("error:", err)
	}
	assets, err := fs.Sub(dist, "assets")
	if err != nil {
		log.Fatalln("error:", err)
	}

	s.StaticFS("/assets/", http.FS(assets))

	s.GET("/", func(c *gin.Context) {
		c.FileFromFS("/", http.FS(dist))
	})
}

// todo 页面如何生成呢？应当支持 SEO 并且能使用 JSON api 获取数据，不能纯 SPA
