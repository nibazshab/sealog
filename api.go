package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

type resource struct {
	Topic *Topic  `json:"topic"`
	Posts []*Post `json:"posts"`
}

func newDiscussion(c *gin.Context) {
	type payload struct {
		Title   string `json:"title"    binding:"required"`
		ModelId int    `json:"model_id" binding:"required"`
		Content string `json:"content"  binding:"required"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		fmt.Println(p)
		responseError(c, err, 400, "payload error")
		return
	}

	topic := Topic{
		Title:   p.Title,
		ModelId: p.ModelId,
	}
	post := Post{
		Content: p.Content,
	}
	err := topic.create(&post)
	if err != nil {
		responseError(c, err, 500, "new discussion error")
		return
	}

	r := resource{
		Topic: &topic,
		Posts: []*Post{&post},
	}
	c.JSON(200, result[resource]{
		Code: 200,
		Msg:  "new discussion success",
		Data: r,
	})
}

func list() {
	// 分页加载，登录状态决定是否归属显示 model.deep = 2  的隐藏帖子
}

func changeUserPassword(c *gin.Context) {
	type payload struct {
		Password string `json:"password" binding:"required"`
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
	err = u.setPassword(p.Password)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}
	c.JSON(200, result[any]{
		Code: 200,
		Msg:  "reset password success",
		Data: nil,
	})
}

func userLogin(c *gin.Context) {
	type payload struct {
		Password string `json:"password" binding:"required"`
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
		if ok {
			c.Set("id", 1)
		} else {
			c.Set("id", 0)
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
