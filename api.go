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

type avData struct {
	Topic Topic  `json:"topic"`
	Posts []Post `json:"posts"`
}

type cvData struct {
	Mode   Mode    `json:"mode"`
	Topics []Topic `json:"topics"`
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

	var topics []Topic
	var err error
	if uid == -1 {
		err = queryTopicsPublic(&topics, urlquery.Offset)
	} else {
		err = queryTopics(&topics, urlquery.Offset)
	}
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	if len(topics) == 21 {
		topics[20] = Topic{Id: -1}
	}

	responseSuccess(c, topics)
}

// api/cv
func getModes(c *gin.Context) {
	uid := c.MustGet("uid").(int)

	var modes []Mode
	var err error
	if uid == -1 {
		err = queryModesPublic(&modes)
	} else {
		err = queryModes(&modes)
	}
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

	topic := Topic{
		Id: aid,
	}
	if err = topic.stat("id"); err != nil {
		responseError(c, err, 404, "not found")
		return
	}

	err = queryTopicAndPostsOnTopic(&topic)
	if err != nil {
		responseError(c, err, 500, "server error")
	}

	if uid == -1 {
		if topic.ModeId == 0 {
			responseError(c, errors.New("access denied"), 404, "not found")
			return
		}
		mode := Mode{
			Id: topic.ModeId,
		}
		err = mode.stat("pub")
		if err != nil {
			responseError(c, err, 500, "server error")
			return
		}
		if mode.Pub == false {
			responseError(c, errors.New("access denied"), 404, "not found")
			return
		}
	}

	var posts []Post
	err = queryTopicAndPostsOnPosts(&posts, aid)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, avData{
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
	if err = mode.stat("pub"); err != nil {
		responseError(c, err, 404, "not found")
		return
	}
	if uid == -1 {
		if mode.Pub == false {
			responseError(c, errors.New("access denied"), 404, "not found")
			return
		}
	}

	err = queryTopicsByModeOnMode(&mode)
	if err != nil {
		responseError(c, err, 500, "server error")
	}

	var topics []Topic
	err = queryTopicsByModeOnTopics(&topics, cid, urlquery.Offset)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	if len(topics) == 21 {
		topics[20] = Topic{Id: -1}
	}

	responseSuccess(c, cvData{
		Mode:   mode,
		Topics: topics,
	})
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

	responseSuccess(c, avData{
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

func queryTopics(dest *[]Topic, offset int) error {
	return db.Order("id DESC").Offset(offset).Limit(21).Find(dest).Error
}

func queryTopicsPublic(dest *[]Topic, offset int) error {
	subQuery := db.Model(&Mode{}).Select("id").Where("pub = ?", true)
	return db.Order("id DESC").Offset(offset).Limit(21).
		Where("mode_id IN (?)", subQuery).Where("mode_id <> 0").Find(dest).Error
}

func queryModes(dest *[]Mode) error {
	return db.Find(dest).Error
}

func queryModesPublic(dest *[]Mode) error {
	return db.Where("pub = ?", true).Find(dest).Error
}

func queryTopicAndPostsOnTopic(dest *Topic) error {
	return db.Model(dest).Where("id = ?", dest.Id).Take(dest).Error
}

func queryTopicAndPostsOnPosts(dest *[]Post, aid int) error {
	return db.Order("floor").Where("topic_id = ?", aid).Find(dest).Error
}

func queryTopicsByModeOnMode(dest *Mode) error {
	return db.Model(dest).Where("id = ?", dest.Id).Take(dest).Error
}

func queryTopicsByModeOnTopics(dest *[]Topic, cid int, offset int) error {
	return db.Order("id DESC").Offset(offset).Limit(21).Where("mode_id = ?", cid).Find(dest).Error
}
