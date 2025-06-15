package main

import (
	"errors"
	"reflect"
	"time"

	"gorm.io/gorm"
)

type (
	// Mode 帖子版块
	Mode struct {
		Id   int    `gorm:"primaryKey"    json:"id"`
		Name string `gorm:"not null"      json:"name"`
		Pub  bool   `gorm:"default:false" json:"pub"`
	}

	// Topic 帖子主题
	Topic struct {
		Id        int       `gorm:"primaryKey"      json:"id"`
		CreatedAt time.Time `gorm:"autoCreateTime"  json:"created_at"`
		Title     string    `gorm:"not null"        json:"title"`
		ModeId    int       `gorm:"index;default:0" json:"mode_id"`
		Floors    int       `gorm:"default:0"       json:"-"`
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
func (m *Mode) create() error {
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
// *m.Name, *m.Pub
func (m *Mode) update(data interface{}) error {
	return db.Model(m).Where("id = ?", m.Id).Omit("id").Updates(data).Error
}

// t.Title, t.ModeId
func (t *Topic) create() error {
	return db.Create(t).Error
}

func (t *Topic) BeforeCreate(*gorm.DB) error {
	mode := Mode{
		Id: t.ModeId,
	}
	return mode.stat("id")
}

// t.Id
func (t *Topic) delete() error {
	return db.Where("id = ?", t.Id).Delete(t).Error
}

func (t *Topic) BeforeDelete(tx *gorm.DB) error {
	return tx.Where("topic_id = ?", t.Id).Delete(&Post{}).Error
}

// t.Id
// *t.Title, *t.ModeId
func (t *Topic) update(data interface{}) error {
	return db.Model(t).Where("id = ?", t.Id).Omit("id", "floors").Updates(data).Error
}

func (t *Topic) BeforeUpdate(tx *gorm.DB) error {
	val := reflect.ValueOf(tx.Statement.Dest)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("not struct")
	}

	mid := val.FieldByName("ModeId")
	if !mid.IsValid() {
		return errors.New("no mode_id")
	}

	if mid.Kind() == reflect.Ptr {
		if mid.IsNil() {
			return nil
		}
		mid = mid.Elem()
	}
	if mid.Kind() != reflect.Int {
		return errors.New("not mode_id")
	}

	mode := Mode{
		Id: int(mid.Int()),
	}
	return mode.stat("id")
}

// p.TopicId, p.Content
func (p *Post) create() error {
	topic := Topic{
		Id: p.TopicId,
	}
	err := topic.stat("floors")
	if err != nil {
		return err
	}

	p.Floor = topic.Floors + 1

	return db.Set("topic_id", p.TopicId).
		Create(p).Error
}

func (p *Post) AfterCreate(tx *gorm.DB) error {
	tid, ok := tx.Get("topic_id")
	if !ok {
		return errors.New("no topic_id")
	}

	return tx.Model(&Topic{}).Session(&gorm.Session{SkipHooks: true}).
		Where("id = ?", tid).Update("floors", gorm.Expr("floors + 1")).Error
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
