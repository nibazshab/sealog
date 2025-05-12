package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type response struct {
	msg string
}

func createTopic(c *gin.Context) {
	var topic Topic
	err := c.ShouldBindJSON(&topic)
	if err != nil {
		c.JSON(http.StatusBadRequest, response{
			msg: err.Error(),
		})
		return
	}
}
