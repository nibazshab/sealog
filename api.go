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

type core interface {
	create(interface{}) error
	update(interface{}) error
	delete() error
}

func getDiscussion(c *gin.Context) {
	id := c.MustGet("id").(int)
	tid, err := strconv.Atoi(c.Param("pid"))
	if err != nil || tid <= 0 {
		responseError(c, err, 404, "not found")
		return
	}

	var data struct {
		Topic
		Deep int
	}

	// SELECT topics.*, modes.deep FROM `topics` LEFT JOIN modes ON topics.mode_id = modes.id WHERE topics.id = 4
	rs := db.Table("topics").Select("topics.*, modes.deep").
		Joins("LEFT JOIN modes ON topics.mode_id = modes.id").
		Where("topics.id = ?", tid).Scan(&data)
	if rs.Error != nil {
		responseError(c, err, 500, "server error")
		return
	}
	if rs.RowsAffected == 0 {
		responseError(c, errors.New("not found"), 404, "not found")
		return
	}

	if id == -1 && (data.Deep == 2 || data.Topic.ModeId == 0) {
		responseError(c, errors.New("access denied"), 404, "not found")
		return
	}

	topic := data.Topic
	var posts []*Post

	// SELECT * FROM `posts` WHERE topic_id = 4 ORDER BY floor
	err = db.Order("floor").Where("topic_id = ?", topic.Id).Find(&posts).Error
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

func getUserId(c *gin.Context) {
	id := c.MustGet("id").(int)

	c.JSON(200, result[int]{
		Code: 200,
		Msg:  "get user id success",
		Data: id,
	})
}

func getCategories(c *gin.Context) {
	id := c.MustGet("id").(int)

	var modes []Mode

	// SELECT * FROM `modes`
	query := db
	if id == -1 {
		// SELECT * FROM `modes` WHERE deep <> 2
		query = query.Where("deep <> ?", 2)
	}
	err := query.Find(&modes).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	c.JSON(200, result[*[]Mode]{
		Code: 200,
		Msg:  "get categories success",
		Data: &modes,
	})
}

func getDiscussionsByCategory(c *gin.Context) {
	id := c.MustGet("id").(int)
	mid, err := strconv.Atoi(c.Param("tid"))
	if err != nil || mid <= 0 {
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
		Id: mid,
	}
	if err = mode.verifyExist(); err != nil {
		responseError(c, err, 404, "not found")
		return
	}

	if id == -1 {
		deep, err := mode.queryDeep()
		if err != nil {
			responseError(c, err, 500, "server error")
			return
		}
		if deep == 2 {
			responseError(c, errors.New("access denied"), 404, "not found")
			return
		}
	}

	var topics []Topic

	// SELECT * FROM `topics` WHERE mode_id = 1 ORDER BY id DESC LIMIT 21
	err = db.Order("id DESC").Offset(urlquery.Offset).Limit(21).
		Where("mode_id = ?", mid).Find(&topics).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	var msg string
	if len(topics) <= 20 {
		msg = "fin"
	} else {
		topics = topics[:20]
		msg = "to be continue"
	}

	c.JSON(200, result[*[]Topic]{
		Code: 200,
		Msg:  msg,
		Data: &topics,
	})
}

func getDiscussions(c *gin.Context) {
	id := c.MustGet("id").(int)
	var urlquery struct {
		Offset int `form:"offset" binding:"min=0"`
	}
	if err := c.ShouldBindQuery(&urlquery); err != nil {
		urlquery.Offset = 0
	}

	var topics []Topic

	// SELECT * FROM `topics` ORDER BY id DESC LIMIT 21
	query := db.Order("id DESC").Offset(urlquery.Offset).Limit(21)
	if id == -1 {
		// SELECT * FROM `topics` WHERE mode_id NOT IN (SELECT `id` FROM `modes` WHERE deep = 2)
		// AND mode_id <> 0 ORDER BY id DESC LIMIT 21
		subQuery := db.Model(&Mode{}).Select("id").Where("deep = ?", 2)
		query = query.Where("mode_id NOT IN (?)", subQuery).Where("mode_id <> 0")
	}
	err := query.Find(&topics).Error
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	var msg string
	if len(topics) <= 20 {
		msg = "fin"
	} else {
		topics = topics[:20]
		msg = "to be continue"
	}

	c.JSON(200, result[*[]Topic]{
		Code: 200,
		Msg:  msg,
		Data: &topics,
	})
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

	mode := Mode{
		Name: payload.Name,
		Deep: payload.Deep,
	}
	publicCreate(c, &mode)
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

	mode := Mode{
		Id: payload.Id,
	}
	newMode := Mode{
		Name: payload.Name,
		Deep: payload.Deep,
	}
	publicUpdate(c, &mode, &newMode)
}

func deleteCategory(c *gin.Context) {
	var payload struct {
		Id int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	mode := Mode{
		Id: payload.Id,
	}
	publicDelete(c, &mode)
}

func createDiscussion(c *gin.Context) {
	var payload struct {
		Title   string `json:"title"    binding:"required"`
		ModeId  int    `json:"mode_id" binding:"required"`
		Content string `json:"content"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responseError(c, err, 400, "payload error")
		return
	}

	topic := Topic{
		Title:  payload.Title,
		ModeId: payload.ModeId,
	}
	post := Post{
		Content: payload.Content,
	}
	r := resource{
		Topic: &topic,
		Posts: []*Post{&post},
	}
	publicCreate(c, &topic, &post, r)
}

func updateDiscussion(c *gin.Context) {
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

	topic := Topic{
		Id: payload.Id,
	}
	newTopic := Topic{
		Title:  payload.Title,
		ModeId: payload.ModeId,
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

func userChangePassword(c *gin.Context) {
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
		Msg:  "change password success",
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
		Msg:  "login success",
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
			c.Set("id", -1)
		}
		c.Next()
	}
}

func protectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.MustGet("id").(int)
		if id == -1 {
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

func publicCreate(c *gin.Context, obj core, conds ...interface{}) {
	var data any
	var r any

	if conds != nil && len(conds) == 2 {
		data = conds[0]
		r = conds[1]
	} else {
		r = obj
	}

	err := obj.create(data)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	c.JSON(200, result[any]{
		Code: 200,
		Msg:  "create success",
		Data: r,
	})
}

func publicUpdate(c *gin.Context, obj core, data interface{}) {
	err := obj.update(data)
	if err != nil {
		responseError(c, err, 500, "server error")
		return
	}

	c.JSON(200, result[core]{
		Code: 200,
		Msg:  "update success",
		Data: obj,
	})
}

func publicDelete(c *gin.Context, obj core) {
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
