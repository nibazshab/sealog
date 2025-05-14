package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func initializeAdminUser() {
	u := User{
		Id: 1,
	}

	// SELECT * FROM `users` WHERE `users`.`id` = 1 ORDER BY `users`.`id` LIMIT 1
	err := db.First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			b := make([]byte, 3)
			rand.Read(b)
			randPassword := hex.EncodeToString(b)

			err = u.setPassword(randPassword)
			if err != nil {
				panic(err)
			}
			log.Println("Password:", randPassword)
		} else {
			panic(err)
		}
	}
}

func (u *User) setPassword(password string) error {
	hash, err := cryptoPassword(password)
	if err != nil {
		return err
	}
	u.Hash = hash

	// INSERT INTO `users` (`hash`,`id`)
	// VALUES ("$2a$10$zqars0w5SAkRBDPnhrBEses1lmy15hfKkGZfO21jx/mi959v3aEfq",1)
	// ON CONFLICT (`id`)
	// DO UPDATE SET `hash`=`excluded`.`hash` RETURNING `id`
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
