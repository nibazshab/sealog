package main

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type (
	// Mode 帖子版块
	Mode struct {
		Id   int    `gorm:"primaryKey" json:"id"`
		Name string `gorm:"not null"   json:"name"`
		Deep int8   `gorm:"default:1"  json:"deep"` // 1 普通, 2 隐藏, 3 公开
	}

	// Topic 帖子主题
	Topic struct {
		Id        int       `gorm:"primaryKey"      json:"id"`
		CreatedAt time.Time `gorm:"autoCreateTime"  json:"created_at"`
		Title     string    `gorm:"not null"        json:"title"`
		ModeId    int       `gorm:"index;default:0" json:"mode_id"`
		Floors    int       `gorm:"default:1"       json:"-"`
	}

	// Post 帖子楼层
	Post struct {
		Id        int       `gorm:"primaryKey"     json:"-"`
		TopicId   int       `gorm:"index;not null" json:"topic_id"`
		Floor     int       `gorm:"index;not null" json:"floor"`
		UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
		Content   string    `gorm:"not null"       json:"content"`
	}
)

// m.Name, m.Deep
func (m *Mode) create(interface{}) error {
	// INSERT INTO `modes` (`name`,`deep`) VALUES ("guest",3) RETURNING `id`
	return db.Create(m).Error
}

// m.Id
func (m *Mode) delete() error {
	// DELETE FROM `modes` WHERE `modes`.`id` = 1
	return db.Delete(m).Error
}

func (m *Mode) BeforeDelete(tx *gorm.DB) error {
	// UPDATE `topics` SET `mode_id`=0 WHERE mode_id = 1
	return tx.Model(&Topic{}).Session(&gorm.Session{SkipHooks: true}).
		Where("mode_id = ?", m.Id).Update("mode_id", 0).Error
}

// m.Id
// m.Id, m.Name, m.Deep
func (m *Mode) update(data interface{}) error {
	// UPDATE `modes` SET `id`=11,`name`="dd" WHERE `id` = 1
	return db.Set("data", data).
		Model(m).Updates(data).Error
}

func (m *Mode) BeforeUpdate(tx *gorm.DB) error {
	if !tx.Statement.Changed("id") {
		return nil
	}

	data, ok := tx.Get("data")
	if !ok {
		return errors.New("no data")
	}
	mode, ok := data.(*Mode)
	if !ok {
		return errors.New("not mode")
	}

	// UPDATE `topics` SET `mode_id`=11 WHERE mode_id = 1
	return tx.Model(&Topic{}).Session(&gorm.Session{SkipHooks: true}).
		Where("mode_id = ?", m.Id).Update("mode_id", mode.Id).Error
}

// t.Title, t.ModeId
// p.Content
func (t *Topic) create(data interface{}) error {
	// INSERT INTO `topics` (`created_at`,`title`,`mode_id`,`floors`) VALUES ("2025-05-16 00:31:07.555","test",1,1) RETURNING `id`
	return db.Set("data", data).
		Create(t).Error
}

func (t *Topic) BeforeCreate(*gorm.DB) error {
	mode := Mode{
		Id: t.ModeId,
	}
	return mode.verifyExist()
}

func (t *Topic) AfterCreate(tx *gorm.DB) error {
	data, ok := tx.Get("data")
	if !ok {
		return errors.New("no data")
	}
	post, ok := data.(*Post)
	if !ok {
		return errors.New("not post")
	}

	post.TopicId = t.Id
	post.Floor = 1

	// INSERT INTO `posts` (`topic_id`,`floor`,`updated_at`,`content`) VALUES (14,1,"2025-05-16 00:31:07.555","Hello World!") RETURNING `id`
	return tx.Model(&Post{}).Session(&gorm.Session{SkipHooks: true}).
		Create(post).Error
}

// t.Id
func (t *Topic) delete() error {
	// DELETE FROM `topics` WHERE `topics`.`id` = 2
	return db.Delete(t).Error
}

func (t *Topic) BeforeDelete(tx *gorm.DB) error {
	// DELETE FROM `posts` WHERE topic_id = 2
	return tx.Where("topic_id = ?", t.Id).Delete(&Post{}).Error
}

// t.Id
// t.Title, t.ModeId
func (t *Topic) update(data interface{}) error {
	// UPDATE `topics` SET `title`="test2",`mode_id`=2 WHERE `id` = 2
	return db.Set("data", data).
		Model(t).Omit("id", "floors").Updates(data).Error
}

func (t *Topic) BeforeUpdate(tx *gorm.DB) error {
	if !tx.Statement.Changed("mode_id") {
		return nil
	}

	data, ok := tx.Get("data")
	if !ok {
		return errors.New("no data")
	}
	topic, ok := data.(*Topic)
	if !ok {
		return errors.New("not topic")
	}

	mode := Mode{
		Id: topic.ModeId,
	}
	return mode.verifyExist()
}

// p.TopicId, p.Content
func (p *Post) create(interface{}) error {
	topic := Topic{
		Id: p.TopicId,
	}
	num, err := topic.queryFloors()
	if err != nil {
		return err
	}

	p.Floor = num + 1

	// INSERT INTO `posts` (`topic_id`,`floor`,`updated_at`,`content`) VALUES (2,8,"2025-05-16 16:37:39.254","Hello.") RETURNING `id`
	return db.Set("data", &topic).
		Create(p).Error
}

func (p *Post) AfterCreate(tx *gorm.DB) error {
	data, ok := tx.Get("data")
	if !ok {
		return errors.New("no data")
	}
	topic, ok := data.(*Topic)
	if !ok {
		return errors.New("not topic")
	}

	// UPDATE `topics` SET `floors`=floors + 1 WHERE `topics`.`id` = 2 AND `topics`.`floors` = 7
	return tx.Model(&Topic{}).Session(&gorm.Session{SkipHooks: true}).
		Where(topic).Update("floors", gorm.Expr("floors + 1")).Error
}

// p.TopicId, p.Floor
func (p *Post) delete() error {
	// DELETE FROM `posts` WHERE topic_id = 2 AND floor = 6
	return db.Where("topic_id = ?", p.TopicId).Where("floor = ?", p.Floor).Delete(p).Error
}

// p.TopicId, p.Floor
// p.Content
func (p *Post) update(data interface{}) error {
	// UPDATE `posts` SET `updated_at`="2025-05-16 18:33:31.041",`content`="" WHERE topic_id = 2 AND floor = 6
	return db.Model(p).Where("topic_id = ?", p.TopicId).Where("floor = ?", p.Floor).
		Select("content").Updates(data).Error
}

// m.Id
func (m *Mode) verifyExist() error {
	var count int64

	// SELECT count(*) FROM `modes` WHERE `modes`.`id` = 1
	err := db.Model(m).Where(m).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("mode not exist")
	}
	return nil
}

// m.Id
func (m *Mode) queryDeep() (int, error) {
	var num int

	// SELECT `deep` FROM `modes` WHERE `modes`.`id` = 2 ORDER BY `modes`.`id` LIMIT 1
	err := db.Select("deep").First(m).Scan(&num).Error
	return num, err
}

// t.Id
func (t *Topic) queryFloors() (int, error) {
	var num int

	// SELECT `floors` FROM `topics` WHERE `topics`.`id` = 2 ORDER BY `topics`.`id` LIMIT 1
	err := db.Select("floors").First(t).Scan(&num).Error
	return num, err
}
