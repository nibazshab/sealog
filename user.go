package main

import (
	"errors"

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
			// 要求用户设置密码
		}
	}
}

func (u *User) setPassword(password string) error {
	hash, err := cryptoPassword(password)
	if err != nil {
		return err
	}
	u.Password = hash

	// INSERT INTO `users` (`password`,`id`)
	// VALUES ("$2a$10$zqars0w5SAkRBDPnhrBEses1lmy15hfKkGZfO21jx/mi959v3aEfq",1)
	// ON CONFLICT (`id`)
	// DO UPDATE SET `password`=`excluded`.`password` RETURNING `id`
	return db.Save(u).Error
}

func (u *User) login(password string) bool {
	// SELECT * FROM `users` WHERE `users`.`id` = 1 ORDER BY `users`.`id` LIMIT 1
	err := db.First(u).Error
	if err != nil {
		return false
	}

	return verifyPassword(password, u.Password)
}
