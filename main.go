package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
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

func cryptoPassword(password string) string {
	salt := rand.Text()
	h := sha256.New()
	h.Write([]byte(password + salt))
	hash := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("$sha256$%s$%s", salt, hash)
}

func verifyPassword(input, password string) bool {
	parts := strings.Split(password, "$")
	if len(parts) != 4 || parts[1] != "sha256" {
		return false
	}
	salt, y := parts[2], parts[3]

	h := sha256.New()
	h.Write([]byte(input + salt))
	x := hex.EncodeToString(h.Sum(nil))

	return y == x
}
