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
		Pub  bool   `                  json:"pub"`
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

// m.Name, m.Pub
func (m *Mode) create(interface{}) error {
	return db.Create(m).Error
}

// m.Id
func (m *Mode) delete() error {
	return db.Where("id = ?", m.Id).Delete(m).Error
}

func (m *Mode) BeforeDelete(tx *gorm.DB) error {
	return tx.Model(&Topic{}).Session(&gorm.Session{SkipHooks: true}).
		Where("mode_id = ?", m.Id).Update("mode_id", 0).Error
}

// m.Id
// m.Id, m.Name, m.Pub
func (m *Mode) update(data interface{}) error {
	return db.Set("data", data).
		Model(m).Where("id = ?", m.Id).Updates(data).Error
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

	return tx.Model(&Topic{}).Session(&gorm.Session{SkipHooks: true}).
		Where("mode_id = ?", m.Id).Update("mode_id", mode.Id).Error
}

// t.Title, t.ModeId
// p.Content
func (t *Topic) create(data interface{}) error {
	return db.Set("data", data).
		Create(t).Error
}

func (t *Topic) BeforeCreate(*gorm.DB) error {
	mode := Mode{
		Id: t.ModeId,
	}
	return mode.stat("id")
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

	return tx.Model(&Post{}).Session(&gorm.Session{SkipHooks: true}).
		Create(post).Error
}

// t.Id
func (t *Topic) delete() error {
	return db.Where("id = ?", t.Id).Delete(t).Error
}

func (t *Topic) BeforeDelete(tx *gorm.DB) error {
	return tx.Where("topic_id = ?", t.Id).Delete(&Post{}).Error
}

// t.Id
// t.Title, t.ModeId
func (t *Topic) update(data interface{}) error {
	return db.Set("data", data).
		Model(t).Where("id = ?", t.Id).Omit("id", "floors").Updates(data).Error
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
	return mode.stat("id")
}

// p.TopicId, p.Content
func (p *Post) create(interface{}) error {
	topic := Topic{
		Id: p.TopicId,
	}
	err := topic.stat("floors")
	if err != nil {
		return err
	}

	p.Floor = topic.Floors + 1

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

	return tx.Model(&Topic{}).Session(&gorm.Session{SkipHooks: true}).
		Where("id = ?", topic.Id).Update("floors", gorm.Expr("floors + 1")).Error
}

// p.TopicId, p.Floor
func (p *Post) delete() error {
	return db.Where("topic_id = ?", p.TopicId).Where("floor = ?", p.Floor).Delete(p).Error
}

// p.TopicId, p.Floor
// p.Content
func (p *Post) update(data interface{}) error {
	return db.Model(p).Where("topic_id = ?", p.TopicId).Where("floor = ?", p.Floor).
		Select("content").Updates(data).Error
}

// m.Id
func (m *Mode) stat(args ...string) error {
	return db.Model(m).Where("id = ?", m.Id).Select(args).Take(m).Error
}

// t.Id
func (t *Topic) stat(args ...string) error {
	return db.Model(t).Where("id = ?", t.Id).Select(args).Take(t).Error
}
