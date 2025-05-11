package main

import (
	"time"

	"golang.org/x/crypto/bcrypt"
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

	// User 角色，游客和管理员, Role: 1 管理员, 0 游客
	User struct {
		Role     int8 `gorm:"primaryKey"`
		Password string
	}

	// JwtSecret Jwt 密钥
	JwtSecret struct {
		Id  int `gorm:"primaryKey"`
		Str string
	}
)

func cryptoPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
