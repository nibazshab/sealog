package main

import (
	"io"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func initializeDbDrive(cfg *config) {
	var err error
	db, err = gorm.Open(
		sqlite.Open(cfg.database+"?_journal=WAL&_vacuum=incremental"),
		&gorm.Config{
			TranslateError: true,
			Logger:         logger.Default.LogMode(logger.Silent),
		})
	if err != nil {
		log.Fatalln("error:", err)
	}

	err = db.AutoMigrate(
		&Mode{}, &Topic{}, &Post{}, &User{}, &Key{},
	)
	if err != nil {
		log.Fatalln("error:", err)
	}
}

func closeDb() {
	coon, err := db.DB()
	if err != nil {
		log.Fatalln("error:", err)
	}
	err = coon.Close()
	if err != nil {
		log.Fatalln("error:", err)
	}
}

func initializeLogDrive(cfg *config) {
	file, err := os.OpenFile(cfg.log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalln("error:", err)
	}
	cfg.w = io.MultiWriter(os.Stdout, file)
	log.SetOutput(cfg.w)
}
