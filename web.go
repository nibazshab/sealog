package main

import (
	"embed"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var web embed.FS

func favicon(c *gin.Context) {
	c.Data(200, "image/x-icon", []byte{})
}

func robots(c *gin.Context) {
	c.Data(200, "text/plain", []byte("User-agent: *\nAllow: /"))
}
