package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户
type User struct {
	Id   int `gorm:"primaryKey"`
	Hash string
}

func initializeAdminUser() {
	u := User{
		Id: 1,
	}

	err := u.getPassword()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			password, err := u.randPassword()
			if err != nil {
				log.Fatalln("error:", err)
			}
			log.Println("default password:", password)
		} else {
			log.Fatalln("error:", err)
		}
	}
}

func resetAdminPassword() (string, error) {
	u := User{
		Id: 1,
	}

	return u.randPassword()
}

func (u *User) randPassword() (string, error) {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalln("error:", err)
	}

	str := hex.EncodeToString(b)
	return str, u.setPassword(str)
}

func (u *User) getPassword() error {
	// SELECT * FROM `users` WHERE `users`.`id` = 1 ORDER BY `users`.`id` LIMIT 1
	return db.First(u).Error
}

func (u *User) setPassword(password string) error {
	hash, err := cryptoPassword(password)
	if err != nil {
		return err
	}
	u.Hash = hash

	// INSERT INTO `users` (`hash`,`id`) VALUES ("$2a$10$zqars0w5SAkRBDPnhrBEses1lmy15hfKkGZfO21jx/mi959v3aEfq",1)
	// ON CONFLICT (`id`) DO UPDATE SET `hash`=`excluded`.`hash` RETURNING `id`
	return db.Save(u).Error
}

func (u *User) login(password string) bool {
	// SELECT * FROM `users` WHERE `users`.`id` = 1 ORDER BY `users`.`id` LIMIT 1
	err := db.First(u).Error
	if err != nil {
		return false
	}

	return verifyPassword(password, u.Hash)
}

func cryptoPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func verifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
