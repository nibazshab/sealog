package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMain(m *testing.M) {
	ex, _ := os.Executable()
	db, _ = gorm.Open(sqlite.Open(filepath.Join(filepath.Dir(ex), "test.db")))
	db.TranslateError = true
	db.Logger = logger.Default.LogMode(logger.Info)

	db.AutoMigrate(&Mode{}, &Topic{}, &Post{})

	println(ex)
	m.Run()
}

func TestNewMode(t *testing.T) {
	fn := func(t *testing.T) {
		const letters = "abcdefghijklmnopqrstuvwxyz"
		b := make([]byte, 5)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}

		mode := Mode{
			Name: string(b),
			Pub:  rand.Intn(2) == 1,
		}

		if err := mode.create(); err != nil {
			t.Error(err)
		}
	}

	fn(t)
	for i := 0; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fn(t)
		})
	}
}

func TestNewTopic(t *testing.T) {
	fn := func(t *testing.T) {
		const letters = "abcdefghijklmnopqrstuvwxyz"
		b := make([]byte, 10)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}

		topic := Topic{
			Title:  string(b),
			ModeId: rand.Intn(10) + 1,
		}

		if err := topic.create(); err != nil {
			t.Error(err)
		}
	}

	fn(t)
	for i := 0; i < 100; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fn(t)
		})
	}
}

func TestNewPost(t *testing.T) {
	fn := func(t *testing.T) {
		const letters = "abcdefghijklmnopqrstuvwxyz"
		b := make([]byte, 10)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}

		post := Post{
			TopicId: rand.Intn(100) + 1,
			Content: string(b),
		}

		if err := post.create(); err != nil {
			t.Error(err)
		}
	}

	fn(t)
	for i := 0; i < 100; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fn(t)
		})
	}
}

func TestUpMode(t *testing.T) {
	mode := Mode{
		Id: 1,
	}

	pub := false
	a := struct {
		Id   int
		Name *string
		Pub  *bool
	}{
		Id:   mode.Id,
		Name: nil,
		Pub:  &pub,
	}

	err := mode.update(a)
	if err != nil {
		t.Error(err)
	}
}

func TestUpTopic(t *testing.T) {
	topic := Topic{
		Id: 99,
	}

	id := 1
	title := "c99"
	b1 := struct {
		ModeId *int
		Title  *string
	}{
		ModeId: &id,
		Title:  &title,
	}
	b2 := Topic{
		ModeId: id,
		Title:  title,
	}
	b3 := map[string]interface{}{
		"mode_id": id,
		"title":   title,
	}
	_, _, _ = b1, b2, b3
	err := topic.update(b1)
	if err != nil {
		t.Error(err)
	}
}

func TestUpPost(t *testing.T) {
	p := Post{
		TopicId: 1,
		Floor:   2,
	}

	b := Post{
		Content: "",
	}

	err := p.update(b)
	if err != nil {
		t.Error(err)
	}
}

func TestDelMode(t *testing.T) {
	mode := Mode{
		Id: 1,
	}

	err := mode.delete()
	if err != nil {
		t.Error(err)
	}
}

func TestDelTopic(t *testing.T) {
	topic := Topic{
		Id: 92,
	}

	err := topic.delete()
	if err != nil {
		t.Error(err)
	}
}

func TestDelPost(t *testing.T) {
	p := Post{
		TopicId: 2,
		Floor:   1,
	}

	err := p.delete()
	if err != nil {
		t.Error(err)
	}
}

func TestTopic_stat(t *testing.T) {
	a := Topic{
		Id: 1,
	}

	err := a.stat("id")
	if err != nil {
		t.Error(err)
	}
}

func TestMode_stat(t *testing.T) {
	mode := Mode{
		Id: 1,
	}

	err := mode.stat("pub")
	if err != nil {
		t.Error(err)
	}
}
