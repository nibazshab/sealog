package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type response struct {
	msg string
}

func createTopic(c *gin.Context) {
	var t Topic
	err := c.ShouldBindJSON(&t) // 结构体需设置 json 可选条目
	if err != nil {
		c.JSON(400, response{
			msg: err.Error(),
		})
		return
	}

	err = t.createNewTopic()
	if err != nil {
		c.JSON(400, response{
			msg: err.Error(),
		})
	}
}

func (t *Topic) createNewTopic() error {
	return db.Create(t).Error
}

func (t *Topic) AfterCreate(tx *gorm.DB) error {
	t.Post.TopicId = t.Id
	t.Post.Floor = 1
	return tx.Create(t.Post).Error
}

func (p *Post) addNewPost() error {
	t := Topic{
		Id: p.TopicId,
	}
	err := t.getTopicFloor()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		p.Floor = t.Floors + 1
		err = tx.Create(p).Error
		if err != nil {
			return err
		}

		return t.addTopicFloor(tx)
	})
}

func (t *Topic) getTopicFloor() error {
	// SELECT `floors` FROM `topics` WHERE `topics`.`id` = 2 ORDER BY `topics`.`id` LIMIT 1
	return db.Model(&Topic{}).Select("floors").First(t).Error
}

func (t *Topic) addTopicFloor(tx *gorm.DB) error {
	// UPDATE `topics` SET `floors`=floors + 1 WHERE `topics`.`id` = 2 AND `topics`.`floors` = 1
	return tx.Model(&Topic{}).Where(t).Update("floors", gorm.Expr("floors + 1")).Error
}
