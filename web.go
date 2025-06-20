package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var web embed.FS

func cacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=2592000")
	}
}

func static(s *gin.RouterGroup) {
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
func renderHtml(h *gin.RouterGroup) {
	h.GET("/", func(c *gin.Context) {})
	h.GET("/av", func(c *gin.Context) {})
	h.GET("/cv", func(c *gin.Context) {})
	h.GET("/av/:aid", func(c *gin.Context) {})
	h.GET("/cv/:cid", func(c *gin.Context) {
		uid := c.MustGet("uid").(int)
		cid, _ := strconv.Atoi(c.Param("cid"))
		var cv resCid
		_, _, _ = queryTopicsByMode(&cv, uid, cid, 0)
		fmt.Println(cv)
	})
}
