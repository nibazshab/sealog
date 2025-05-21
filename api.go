package main

import (
	"errors"
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

func createCategory(c *gin.Context) {
	type payload struct {
		Name string `json:"name" binding:"required"`
		Deep int8   `json:"deep" binding:"required"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	model := Model{
		Name: p.Name,
		Deep: p.Deep,
	}
	publicCreate(c, &model)
}

func updateCategory(c *gin.Context) {
	type payload struct {
		Id   int    `json:"id" binding:"required"`
		Name string `json:"name"`
		Deep int8   `json:"deep"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	if p.Name == "" && p.Deep == 0 {
		responseError(c, errors.New("missing value"), 400, "payload error")
		return
	}

	model := Model{
		Id: p.Id,
	}
	newModel := Model{
		Name: p.Name,
		Deep: p.Deep,
	}
	publicUpdate(c, &model, &newModel)
}

func deleteCategory(c *gin.Context) {
	type payload struct {
		Id int `json:"id" binding:"required"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	model := Model{
		Id: p.Id,
	}
	publicDelete(c, &model)
}

func createDiscussion(c *gin.Context) {
	type payload struct {
		Title   string `json:"title"    binding:"required"`
		ModelId int    `json:"model_id" binding:"required"`
		Content string `json:"content"  binding:"required"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
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
		responseError(c, err, 500, "server error")
		return
	}

	r := resource{
		Topic: &topic,
		Posts: []*Post{&post},
	}
	c.JSON(200, result[resource]{
		Code: 200,
		Msg:  "create success",
		Data: r,
	})
}

func updateDiscussion(c *gin.Context) {
	type payload struct {
		Id      int    `json:"id" binding:"required"`
		Title   string `json:"title"`
		ModelId int    `json:"model_id"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	if p.Title == "" && p.ModelId == 0 {
		responseError(c, errors.New("missing value"), 400, "payload error")
		return
	}

	topic := Topic{
		Id: p.Id,
	}
	newTopic := Topic{
		Title:   p.Title,
		ModelId: p.ModelId,
	}
	publicUpdate(c, &topic, &newTopic)
}

func deleteDiscussion(c *gin.Context) {
	type payload struct {
		Id int `json:"id" binding:"required"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	topic := Topic{
		Id: p.Id,
	}
	publicDelete(c, &topic)
}

func createComment(c *gin.Context) {
	type payload struct {
		TopicId int    `json:"topic_id" binding:"required"`
		Content string `json:"content"  binding:"required"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	post := Post{
		TopicId: p.TopicId,
		Content: p.Content,
	}
	publicCreate(c, &post)
}

func updateComment(c *gin.Context) {
	type payload struct {
		TopicId int    `json:"topic_id" binding:"required"`
		Floor   int    `json:"floor"    binding:"required"`
		Content string `json:"content"  binding:"required"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	post := Post{
		TopicId: p.TopicId,
		Floor:   p.Floor,
	}
	newPost := Post{
		Content: p.Content,
	}
	publicUpdate(c, &post, &newPost)
}

func deleteComment(c *gin.Context) {
	type payload struct {
		TopicId int `json:"topic_id" binding:"required"`
		Floor   int `json:"floor"    binding:"required"`
	}
	var p payload
	if err := c.ShouldBindJSON(&p); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	post := Post{
		TopicId: p.TopicId,
		Floor:   p.Floor,
	}
	publicDelete(c, &post)
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
		Msg:  "change user password success",
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

type coreTypes interface {
	*Model | *Topic | *Post
}

type coreCreate interface {
	create() error
}

type coreUpdate[T coreTypes] interface {
	update(T) error
}

type coreDelete interface {
	delete() error
}

func publicCreate(c *gin.Context, obj coreCreate) {
	err := obj.create()
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	c.JSON(200, result[coreCreate]{
		Code: 200,
		Msg:  "create success",
		Data: obj,
	})
}

func publicUpdate[T coreTypes](c *gin.Context, obj coreUpdate[T], new T) {
	err := obj.update(new)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	c.JSON(200, result[coreUpdate[T]]{
		Code: 200,
		Msg:  "update success",
		Data: obj,
	})
}

func publicDelete(c *gin.Context, obj coreDelete) {
	err := obj.delete()
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	c.JSON(200, result[any]{
		Code: 200,
		Msg:  "delete success",
		Data: nil,
	})
}
