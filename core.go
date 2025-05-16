package main

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type (
	// Model 帖子版块, Deep: 1 游客可以查看, 2 游客不可查看, 3 游客可以发帖
	Model struct {
		Id   int    `gorm:"primaryKey" json:"id"`
		Name string `json:"name"`
		Deep int8   `gorm:"default:1" json:"deep"`
	}

	// Topic 帖子主题
	Topic struct {
		Id        int       `gorm:"primaryKey"`
		CreatedAt time.Time `gorm:"autoCreateTime"`
		Title     string    `json:"title"`
		ModelId   int       `gorm:"index" json:"model_id"`
		Floors    int       `gorm:"default:1"`
	}

	// Post 帖子楼层
	Post struct {
		Id        int       `gorm:"primaryKey"`
		TopicId   int       `gorm:"index" json:"topic_id"`
		Floor     int       `gorm:"index" json:"floor"`
		UpdatedAt time.Time `gorm:"autoUpdateTime"`
		Content   string    `json:"content"`
	}
)

func (m *Model) create() error {
	// INSERT INTO `models` (`name`,`deep`) VALUES ("测试",1) RETURNING `id`
	return db.Create(m).Error
}

func (m *Model) delete() error {
	// DELETE FROM `models` WHERE `models`.`id` = 1
	return db.Delete(m).Error
}

func (m *Model) BeforeDelete(tx *gorm.DB) error {
	// UPDATE `topics` SET `model_id`=0 WHERE model_id = 1
	return tx.Model(&Topic{}).Where("model_id = ?", m.Id).Update("model_id", 0).Error
}

func (m *Model) update(new *Model) error {
	// UPDATE `models` SET `id`=11,`name`="dd" WHERE `id` = 1
	return db.Set("newId", new.Id).Model(m).Updates(new).Error
}

func (m *Model) BeforeUpdate(tx *gorm.DB) error {
	if !tx.Statement.Changed("id") {
		return nil
	}

	newId, ok := tx.Get("newId")
	if !ok {
		return errors.New("no newId")
	}

	n, ok := newId.(int)
	if !ok {
		return errors.New("not newId")
	}

	// UPDATE `topics` SET `model_id`=11 WHERE model_id = 1
	return tx.Model(&Topic{}).Where("model_id = ?", m.Id).Update("model_id", n).Error
}

func (m *Model) verifyExist() error {
	var count int64

	// SELECT count(*) FROM `models` WHERE `models`.`id` = 1
	err := db.Model(m).Where(m).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("model not exist")
	}
	return nil
}

func (t *Topic) create(p *Post) error {
	// INSERT INTO `topics` (`created_at`,`title`,`model_id`,`floors`) VALUES ("2025-05-16 00:31:07.555","test",1,1) RETURNING `id`
	return db.Set("post", p).Create(t).Error
}

func (t *Topic) BeforeCreate(*gorm.DB) error {
	m := Model{
		Id: t.ModelId,
	}
	return m.verifyExist()
}

func (t *Topic) AfterCreate(tx *gorm.DB) error {
	post, ok := tx.Get("post")
	if !ok {
		return errors.New("no post")
	}

	p, ok := post.(*Post)
	if !ok {
		return errors.New("not post")
	}

	p.TopicId = t.Id
	p.Floor = 1

	// INSERT INTO `posts` (`topic_id`,`floor`,`updated_at`,`content`) VALUES (14,1,"2025-05-16 00:31:07.555","Hello World!") RETURNING `id`
	return tx.Create(p).Error
}

func (t *Topic) countFloor() error {
	// SELECT `floors` FROM `topics` WHERE `topics`.`id` = 2 ORDER BY `topics`.`id` LIMIT 1
	return db.Model(&Topic{}).Select("floors").First(t).Error
}

func (p *Post) additional() error {
	t := Topic{
		Id: p.TopicId,
	}
	err := t.countFloor()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		p.Floor = t.Floors + 1

		// INSERT INTO `posts` (`topic_id`,`floor`,`updated_at`,`content`) VALUES (2,8,"2025-05-16 16:37:39.254","Hello.") RETURNING `id`
		err = tx.Create(p).Error
		if err != nil {
			return err
		}

		// UPDATE `topics` SET `floors`=floors + 1 WHERE `topics`.`id` = 2 AND `topics`.`floors` = 7
		return tx.Model(&Topic{}).Where(&t).Update("floors", gorm.Expr("floors + 1")).Error
	})
}
