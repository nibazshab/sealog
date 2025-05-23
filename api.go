package main

import (
	"errors"
	"log"
	"strconv"

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

type (
	coreTypes interface {
		*Model | *Topic | *Post
	}

	coreCreate interface {
		create() error
	}

	coreUpdate[T coreTypes] interface {
		update(T) error
	}

	coreDelete interface {
		delete() error
	}
)

func getDiscussion(c *gin.Context) {
	id := c.MustGet("id").(int)
	tid, err := strconv.Atoi(c.Param("tid"))
	if err != nil {
		responseError(c, err, 404, "not found")
		return
	}

	var data struct {
		Topic
		Deep int
	}

	// SELECT topics.*, models.deep FROM `topics` LEFT JOIN models ON topics.model_id = models.id WHERE topics.id = 4
	err = db.Table("topics").Select("topics.*, models.deep").
		Joins("LEFT JOIN models ON topics.model_id = models.id").
		Where("topics.id = ?", tid).Scan(&data).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	if id == 0 && (data.Deep == 2 || data.Topic.ModelId == 0) {
		responseError(c, errors.New("access denied"), 404, "not found")
		return
	}

	topic := data.Topic
	var posts []*Post

	// SELECT * FROM `posts` WHERE topic_id = 4
	err = db.Where("topic_id = ?", topic.Id).Find(&posts).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	r := resource{
		Topic: &topic,
		Posts: posts,
	}
	c.JSON(200, result[resource]{
		Code: 200,
		Msg:  "get discussion success",
		Data: r,
	})
}

func getUserInformation(c *gin.Context) {
	id := c.MustGet("id").(int)

	c.JSON(200, result[int]{
		Code: 200,
		Msg:  "get user information success",
		Data: id,
	})
}

func getCategories(c *gin.Context) {
	id := c.MustGet("id").(int)

	// SELECT * FROM `models` WHERE deep <> 2
	//
	// SELECT * FROM `models`
	query := db
	if id == 0 {
		query = query.Where("deep <> ?", 2)
	}
	var models []Model
	err := query.Find(&models).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	c.JSON(200, result[*[]Model]{
		Code: 200,
		Msg:  "get categories success",
		Data: &models,
	})
}

func getDiscussions(c *gin.Context) {
	id := c.MustGet("id").(int)
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// SELECT * FROM `topics` WHERE model_id NOT IN (SELECT `id` FROM `models` WHERE deep = 2)
	// AND model_id <> 0 ORDER BY id DESC LIMIT 21
	//
	// SELECT * FROM `topics` ORDER BY id DESC LIMIT 21
	query := db.Order("id DESC").Offset(offset).Limit(21)
	if id == 0 {
		subQuery := db.Model(&Model{}).Select("id").Where("deep = ?", 2)
		query = query.Where("model_id NOT IN (?)", subQuery).Where("model_id <> 0")
	}
	var topics []Topic
	err = query.Find(&topics).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	var message string
	end := len(topics) <= 20
	if end {
		message = "fin"
	} else {
		topics = topics[:20]
		message = "to be continue"
	}

	c.JSON(200, result[*[]Topic]{
		Code: 200,
		Msg:  message,
		Data: &topics,
	})
}

func responseList(c *gin.Context, conds ...interface{}) {
	// 初次打开网页时的初始化返回应当包括：
	// []model, []topic, user 信息
}

func createCategory(c *gin.Context) {
	var payload struct {
		Name string `json:"name" binding:"required"`
		Deep int8   `json:"deep" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	model := Model{
		Name: payload.Name,
		Deep: payload.Deep,
	}
	publicCreate(c, &model)
}

func updateCategory(c *gin.Context) {
	var payload struct {
		Id   int    `json:"id" binding:"required"`
		Name string `json:"name"`
		Deep int8   `json:"deep"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	if payload.Name == "" && payload.Deep == 0 {
		responseError(c, errors.New("missing value"), 400, "payload error")
		return
	}

	model := Model{
		Id: payload.Id,
	}
	newModel := Model{
		Name: payload.Name,
		Deep: payload.Deep,
	}
	publicUpdate(c, &model, &newModel)
}

func deleteCategory(c *gin.Context) {
	var payload struct {
		Id int `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	model := Model{
		Id: payload.Id,
	}
	publicDelete(c, &model)
}

func createDiscussion(c *gin.Context) {
	var payload struct {
		Title   string `json:"title"    binding:"required"`
		ModelId int    `json:"model_id" binding:"required"`
		Content string `json:"content"  binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	topic := Topic{
		Title:   payload.Title,
		ModelId: payload.ModelId,
	}
	post := Post{
		Content: payload.Content,
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
	var payload struct {
		Id      int    `json:"id" binding:"required"`
		Title   string `json:"title"`
		ModelId int    `json:"model_id"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	if payload.Title == "" && payload.ModelId == 0 {
		responseError(c, errors.New("missing value"), 400, "payload error")
		return
	}

	topic := Topic{
		Id: payload.Id,
	}
	newTopic := Topic{
		Title:   payload.Title,
		ModelId: payload.ModelId,
	}
	publicUpdate(c, &topic, &newTopic)
}

func deleteDiscussion(c *gin.Context) {
	var payload struct {
		Id int `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	topic := Topic{
		Id: payload.Id,
	}
	publicDelete(c, &topic)
}

func createComment(c *gin.Context) {
	var payload struct {
		TopicId int    `json:"topic_id" binding:"required"`
		Content string `json:"content"  binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	post := Post{
		TopicId: payload.TopicId,
		Content: payload.Content,
	}
	publicCreate(c, &post)
}

func updateComment(c *gin.Context) {
	var payload struct {
		TopicId int    `json:"topic_id" binding:"required"`
		Floor   int    `json:"floor"    binding:"required"`
		Content string `json:"content"  binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	post := Post{
		TopicId: payload.TopicId,
		Floor:   payload.Floor,
	}
	newPost := Post{
		Content: payload.Content,
	}
	publicUpdate(c, &post, &newPost)
}

func deleteComment(c *gin.Context) {
	var payload struct {
		TopicId int `json:"topic_id" binding:"required"`
		Floor   int `json:"floor"    binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	post := Post{
		TopicId: payload.TopicId,
		Floor:   payload.Floor,
	}
	publicDelete(c, &post)
}

func changeUserPassword(c *gin.Context) {
	var payload struct {
		Password string `json:"password" binding:"required"`
	}

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	u := User{
		Id: 1,
	}
	err = u.setPassword(payload.Password)
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
	var payload struct {
		Password string `json:"password" binding:"required"`
	}

	err := c.ShouldBindJSON(&payload)
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

	ok := u.login(payload.Password)
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

func protectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.MustGet("id").(int)
		if id == 0 {
			c.AbortWithStatusJSON(200, result[any]{
				Code: 403,
				Msg:  "access denied",
				Data: nil,
			})
			return
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
