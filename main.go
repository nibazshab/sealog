package main

import (
	"github.com/gin-gonic/gin"
)

func run() {
	r := gin.Default()
	r.Run(":8080")
}
