package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

type (
	// Topic 帖子主题
	Topic struct {
		Id        int       `gorm:"primaryKey"`
		CreatedAt time.Time `gorm:"autoCreateTime"`
		Title     string
		Tag       int `gorm:"index"`
		Floors    int `gorm:"default:1"`
	}

	// Post 帖子楼层
	Post struct {
		Id        int `gorm:"primaryKey"`
		TopicId   int `gorm:"index"`
		Floor     int
		UpdatedAt time.Time `gorm:"autoUpdateTime"`
		Content   string
	}

	// Tag 帖子版块, Deep: 1 游客可以查看, 2 游客不可查看, 3 游客可以发帖
	Tag struct {
		Id   int `gorm:"primaryKey"`
		Name string
		Deep int8 `gorm:"default:1"`
	}

	// User 用户
	User struct {
		Id   int `gorm:"primaryKey"`
		Hash string
	}

	// Key 签名密钥
	Key struct {
		Id  int `gorm:"primaryKey"`
		Str string
	}
)

func run() {
	r := gin.Default()
	r.Run(":8080")
}
