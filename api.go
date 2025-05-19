package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

type result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

type recvTopicBody struct {
	Topic
	Post
}

// recv json like
//
//	{
//		"title": "test",
//		"model_id": 1,
//		"content": "Hello World!"
//	}

func login(c *gin.Context) {
	type payload struct {
		Password string `json:"password"`
	}
	var p payload
	err := c.ShouldBindJSON(&p)
	if err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	u := User{
		Id: 1,
	}
	err = u.getPassword()
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	ok := u.login(p.Password)
	if !ok {
		responseError(c, err, 401, "password error")
		return
	}
	token, err := encodeToken()
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}
	c.JSON(200, result[string]{
		Code: 200,
		Msg:  "token",
		Data: token,
	})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")

		ok := decodeToken(token)
		if !ok {
			c.Set("id", 0)
		} else {
			c.Set("id", 1)
		}

		c.Next()
	}
}

func responseError(c *gin.Context, err error, code int, msg string) {
	log.Println("error:", err)

	c.JSON(200, result[any]{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
