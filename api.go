package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

type result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

type (
	resAid struct {
		Topic Topic  `json:"topic"`
		Posts []Post `json:"posts"`
	}

	resCid struct {
		Mode   Mode    `json:"mode"`
		Topics []Topic `json:"topics"`
	}
)

type core interface {
	create() error
	update(interface{}) error
	delete() error
}

// api/search
func getTopicsBySearch(c *gin.Context) {
	uid := c.MustGet("uid").(int)
	var urlquery struct {
		Q string `form:"q"`
	}
	if err := c.ShouldBindQuery(&urlquery); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}
	urlquery.Q = strings.TrimSpace(urlquery.Q)
	if utf8.RuneCountInString(urlquery.Q) < 3 {
		responseError(c, errors.New("too few characters"), 400, "too few characters")
		return
	}

	var topics []Topic
	err, code, msg := queryTopicsBySearch(&topics, uid, urlquery.Q)
	if err != nil {
		responseError(c, err, code, msg)
		return
	}

	responseSuccess(c, topics)
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
	err := queryTopics(&topics, uid, urlquery.Offset)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, topics)
}

// api/cv
func getModes(c *gin.Context) {
	uid := c.MustGet("uid").(int)

	var modes []Mode
	err := queryModes(&modes, uid)
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

	var res resAid
	err, code, msg := queryTopicAndPosts(&res, uid, aid)
	if err != nil {
		responseError(c, err, code, msg)
		return
	}

	responseSuccess(c, res)
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

	var res resCid
	err, code, msg := queryTopicsByMode(&res, uid, cid, urlquery.Offset)
	if err != nil {
		responseError(c, err, code, msg)
		return
	}

	responseSuccess(c, res)
}

func queryTopicsBySearch(dest *[]Topic, uid int, chars string) (error, int, string) {
	var idx []int
	err := db.Model(&Post{}).Where("content LIKE ?", "%"+chars+"%").Select("topic_id").Scan(&idx).Error
	if err != nil {
		return err, 500, "server error"
	}

	if len(idx) == 0 {
		dest = nil
		return nil, 200, "success"
	}

	idy := make([]int, 0, len(idx))
	f := make(map[int]bool)
	for _, x := range idx {
		if !f[x] {
			f[x] = true
			idy = append(idy, x)
		}
	}

	if uid == -1 {
		subQuery := db.Model(&Mode{}).Select("id").Where("pub = ?", true)
		err = db.Order("id DESC").Where("id IN (?)", idy).
			Where("mode_id IN (?)", subQuery).Where("mode_id <> 0").
			Select("id", "title", "mode_id").Find(dest).Error
	} else {
		err = db.Order("id DESC").Where("id IN (?)", idy).Select("id", "title", "mode_id").Find(dest).Error
	}
	if err != nil {
		return err, 500, "server error"
	}

	return nil, 200, "success"
}

func queryTopics(dest *[]Topic, uid int, offset int) error {
	var err error
	if uid == -1 {
		subQuery := db.Model(&Mode{}).Select("id").Where("pub = ?", true)
		err = db.Order("id DESC").Offset(offset).Limit(21).
			Where("mode_id IN (?)", subQuery).Where("mode_id <> 0").Find(dest).Error
	} else {
		err = db.Order("id DESC").Offset(offset).Limit(21).Find(dest).Error
	}
	if err != nil {
		return err
	}

	if len(*dest) == 21 {
		(*dest)[20] = Topic{Id: -1}
	}

	return nil
}

func queryModes(dest *[]Mode, uid int) error {
	var err error
	if uid == -1 {
		err = db.Where("pub = ?", true).Find(dest).Error
	} else {
		err = db.Find(dest).Error
	}

	return err
}

func queryTopicAndPosts(dest *resAid, uid int, aid int) (error, int, string) {
	topic := Topic{
		Id: aid,
	}
	err := topic.stat("id")
	if err != nil {
		return err, 404, "not found"
	}

	err = db.Model(&topic).Where("id = ?", aid).Take(&topic).Error
	if err != nil {
		return err, 500, "server error"
	}

	if uid == -1 {
		if topic.ModeId == 0 {
			return errors.New("access denied"), 404, "not found"
		}
		mode := Mode{
			Id: topic.ModeId,
		}
		err = mode.stat("pub")
		if err != nil {
			return err, 500, "server error"
		}
		if mode.Pub == false {
			return errors.New("access denied"), 404, "not found"
		}
	}

	var posts []Post
	err = db.Order("floor").Where("topic_id = ?", aid).Find(&posts).Error
	if err != nil {
		return err, 500, "server error"
	}

	dest.Topic = topic
	dest.Posts = posts

	return nil, 200, ""
}

func queryTopicsByMode(dest *resCid, uid int, cid int, offset int) (error, int, string) {
	mode := Mode{
		Id: cid,
	}
	err := mode.stat("pub")
	if err != nil {
		return err, 404, "not found"
	}
	if uid == -1 && mode.Pub == false {
		return errors.New("access denied"), 404, "not found"
	}

	err = db.Model(&mode).Where("id = ?", cid).Take(&mode).Error
	if err != nil {
		return err, 500, "server error"
	}

	var topics []Topic
	err = db.Order("id DESC").Offset(offset).Limit(21).Where("mode_id = ?", cid).Find(&topics).Error
	if err != nil {
		return err, 500, "server error"
	}

	if len(topics) == 21 {
		topics[20] = Topic{Id: -1}
	}

	dest.Mode = mode
	dest.Topics = topics

	return nil, 200, ""
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

	err := coreCreate(&obj)
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
		Id   int     `json:"id" binding:"required"`
		Name *string `json:"name"`
		Pub  *bool   `json:"pub"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}
	if payload.Name == nil && payload.Pub == nil {
		responseError(c, errors.New("missing value"), 400, "payload error")
		return
	}

	obj := Mode{
		Id: payload.Id,
	}

	err := coreUpdate(&obj, payload)
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

	err := coreCreate(&obj)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	data := Post{
		TopicId: obj.Id,
		Content: payload.Content,
	}

	err = coreCreate(&data)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, resAid{
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
		Id     int     `json:"id" binding:"required"`
		Title  *string `json:"title"`
		ModeId *int    `json:"mode_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	if payload.Title == nil && payload.ModeId == nil {
		responseError(c, errors.New("missing value"), 400, "payload error")
		return
	}

	obj := Topic{
		Id: payload.Id,
	}

	err := coreUpdate(&obj, payload)
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

	err := coreCreate(&obj)
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

	err := coreUpdate(&obj, payload)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	responseSuccess(c, obj)
}

// api/space
func getAuthStat(c *gin.Context) {
	uid := c.MustGet("uid").(int)
	var stat bool

	if uid == 1 {
		stat = true
	}

	responseSuccess(c, stat)
}

// api/login
func verifyAuthKey(c *gin.Context) {
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
func changeAuthKey(c *gin.Context) {
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

func coreCreate(obj core) error {
	return obj.create()
}

func coreUpdate(obj core, data interface{}) error {
	return obj.update(data)
}

func coreDelete(obj core) error {
	return obj.delete()
}
