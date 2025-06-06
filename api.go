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
	Topic Topic  `json:"topic"`
	Posts []Post `json:"posts"`
}

type core interface {
	create(interface{}) error
	update(interface{}) error
	delete() error
}

// api/av
func getTopics(c *gin.Context) {
	uid := c.MustGet("uid").(int)
	var urlquery struct {
		Offset int `form:"offset" binding:"min=0"`
	}
	if err := c.ShouldBindQuery(&urlquery); err != nil {
		urlquery.Offset = 0
	}

	var topics []*Topic

	// SELECT * FROM `topics` ORDER BY id DESC LIMIT 21
	query := db.Order("id DESC").Offset(urlquery.Offset).Limit(21)
	if uid == -1 {
		// SELECT * FROM `topics` WHERE mode_id NOT IN (SELECT `id` FROM `modes` WHERE pub = false)
		// AND mode_id <> 0 ORDER BY id DESC LIMIT 21
		subQuery := db.Model(&Mode{}).Select("id").Where("pub = ?", false)
		query = query.Where("mode_id NOT IN (?)", subQuery).Where("mode_id <> 0")
	}
	err := query.Find(&topics).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	if len(topics) == 21 {
		topics[20] = &Topic{Id: -1}
	}

	responseSuccess(c, topics)
}

// api/cv
func getModes(c *gin.Context) {
	uid := c.MustGet("uid").(int)

	var modes []*Mode

	// SELECT * FROM `modes`
	query := db
	if uid == -1 {
		// SELECT * FROM `modes` WHERE pub <> false
		query = query.Where("pub <> ?", false)
	}
	err := query.Find(&modes).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, modes)
}

// api/av/:aid
func getTopicAndPosts(c *gin.Context) {
	uid := c.MustGet("uid").(int)
	aid, err := strconv.Atoi(c.Param("aid"))
	if err != nil || aid <= 0 {
		responseError(c, err, 404, "not found")
		return
	}

	var data struct {
		Topic
		Pub bool
	}

	// SELECT topics.*, modes.pub FROM `topics` LEFT JOIN modes ON topics.mode_id = modes.id WHERE topics.id = 4
	rs := db.Table("topics").Select("topics.*, modes.pub").
		Joins("LEFT JOIN modes ON topics.mode_id = modes.id").
		Where("topics.id = ?", aid).Scan(&data)
	if rs.Error != nil {
		responseError(c, rs.Error, 500, "server error")
		return
	}
	if rs.RowsAffected == 0 {
		responseError(c, errors.New("not found"), 404, "not found")
		return
	}

	if uid == -1 && (data.Pub == false || data.Topic.ModeId == 0) {
		responseError(c, errors.New("access denied"), 404, "not found")
		return
	}

	topic := data.Topic
	var posts []Post

	// SELECT * FROM `posts` WHERE topic_id = 4 ORDER BY floor
	err = db.Order("floor").Where("topic_id = ?", topic.Id).Find(&posts).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, resource{
		Topic: topic,
		Posts: posts,
	})
}

// api/cv/:cid
func getTopicsByMode(c *gin.Context) {
	uid := c.MustGet("uid").(int)
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil || cid <= 0 {
		responseError(c, err, 404, "not found")
		return
	}
	var urlquery struct {
		Offset int `form:"offset" binding:"min=0"`
	}
	if err = c.ShouldBindQuery(&urlquery); err != nil {
		urlquery.Offset = 0
	}

	mode := Mode{
		Id: cid,
	}
	if err = mode.verifyExist(); err != nil {
		responseError(c, err, 404, "not found")
		return
	}

	if uid == -1 {
		pub, err := mode.queryPublic()
		if err != nil {
			responseError(c, err, 500, "server error")
			return
		}
		if pub == false {
			responseError(c, errors.New("access denied"), 404, "not found")
			return
		}
	}

	var topics []*Topic

	// SELECT * FROM `topics` WHERE mode_id = 1 ORDER BY id DESC LIMIT 21
	err = db.Order("id DESC").Offset(urlquery.Offset).Limit(21).
		Where("mode_id = ?", cid).Find(&topics).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	if len(topics) == 21 {
		topics[20] = &Topic{Id: -1}
	}

	responseSuccess(c, topics)
}

// api/cv/create
func createMode(c *gin.Context) {
	var payload struct {
		Name string `json:"name" binding:"required"`
		Pub  bool   `json:"pub"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	obj := Mode{
		Name: payload.Name,
		Pub:  payload.Pub,
	}

	err := coreCreate(&obj, nil)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, obj)
}

// api/cv/delete
func deleteMode(c *gin.Context) {
	var payload struct {
		Id int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	obj := Mode{
		Id: payload.Id,
	}

	err := coreDelete(&obj)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, (*struct{})(nil))
}

// api/cv/update
func updateMode(c *gin.Context) {
	var payload struct {
		Id   int    `json:"id" binding:"required"`
		Name string `json:"name"`
		Pub  bool   `json:"pub"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	if payload.Name == "" {
		responseError(c, errors.New("missing value"), 400, "payload error")
		return
	}

	obj := Mode{
		Id: payload.Id,
	}
	data := Mode{
		Name: payload.Name,
		Pub:  payload.Pub,
	}

	err := coreUpdate(&obj, &data)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, obj)
}

// api/av/create
func createTopic(c *gin.Context) {
	var payload struct {
		Title   string `json:"title"    binding:"required"`
		ModeId  int    `json:"mode_id" binding:"required"`
		Content string `json:"content"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	obj := Topic{
		Title:  payload.Title,
		ModeId: payload.ModeId,
	}
	data := Post{
		Content: payload.Content,
	}

	err := coreCreate(&obj, &data)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, resource{
		Topic: obj,
		Posts: []Post{data},
	})
}

// api/av/delete
func deleteTopic(c *gin.Context) {
	var payload struct {
		Id int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	obj := Topic{
		Id: payload.Id,
	}

	err := coreDelete(&obj)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, (*struct{})(nil))
}

// api/av/update
func updateTopic(c *gin.Context) {
	var payload struct {
		Id     int    `json:"id" binding:"required"`
		Title  string `json:"title"`
		ModeId int    `json:"mode_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	if payload.Title == "" && payload.ModeId == 0 {
		responseError(c, errors.New("missing value"), 400, "payload error")
		return
	}

	obj := Topic{
		Id: payload.Id,
	}
	data := Topic{
		Title:  payload.Title,
		ModeId: payload.ModeId,
	}

	err := coreUpdate(&obj, &data)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, obj)
}

// api/fl/create
func createPost(c *gin.Context) {
	var payload struct {
		TopicId int    `json:"topic_id" binding:"required"`
		Content string `json:"content"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	obj := Post{
		TopicId: payload.TopicId,
		Content: payload.Content,
	}

	err := coreCreate(&obj, nil)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, obj)
}

// api/fl/delete
func deletePost(c *gin.Context) {
	var payload struct {
		TopicId int `json:"topic_id" binding:"required"`
		Floor   int `json:"floor"    binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	obj := Post{
		TopicId: payload.TopicId,
		Floor:   payload.Floor,
	}

	err := coreDelete(&obj)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, (*struct{})(nil))
}

// api/fl/update
func updatePost(c *gin.Context) {
	var payload struct {
		TopicId int    `json:"topic_id" binding:"required"`
		Floor   int    `json:"floor"    binding:"required"`
		Content string `json:"content"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	obj := Post{
		TopicId: payload.TopicId,
		Floor:   payload.Floor,
	}
	data := Post{
		Content: payload.Content,
	}

	err := coreUpdate(&obj, &data)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, obj)
}

// api/uid
func getAuthUid(c *gin.Context) {
	uid := c.MustGet("uid").(int)

	responseSuccess(c, uid)
}

// api/auth/login
func loginAuth(c *gin.Context) {
	var payload struct {
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	hash, err := getAuthHash()
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	ok := verifyPassword(hash, payload.Password)
	if !ok {
		responseError(c, errors.New("password error"), 401, "password error")
		return
	}
	token, err := encodeToken()
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, token)
}

// api/auth/change
func changeAuth(c *gin.Context) {
	var payload struct {
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	err = addAuthHash(payload.Password)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, (*struct{})(nil))
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		ok := decodeToken(token)
		if ok {
			c.Set("uid", 1)
		} else {
			c.Set("uid", -1)
		}
		c.Next()
	}
}

func protectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.MustGet("uid").(int)
		if uid == -1 {
			c.AbortWithStatusJSON(403, result[*struct{}]{
				Code: 403,
				Msg:  "access denied",
				Data: nil,
			})
			return
		}
		c.Next()
	}
}

func responseSuccess[T any](c *gin.Context, data T) {
	c.JSON(200, result[T]{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func responseError(c *gin.Context, err error, code int, msg string) {
	log.Println("error:", err)

	c.JSON(code, result[*struct{}]{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func coreCreate(obj core, data interface{}) error {
	return obj.create(data)
}

func coreUpdate(obj core, data interface{}) error {
	return obj.update(data)
}

func coreDelete(obj core) error {
	return obj.delete()
}
