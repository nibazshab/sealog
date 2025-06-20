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
	db, _ = gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "test.db")))
	db.TranslateError = true
	db.Logger = logger.Default.LogMode(logger.Info)

	_ = db.AutoMigrate(&Mode{}, &Topic{}, &Post{})
	m.Run()
}

func TestMode_create(t *testing.T) {
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

func TestTopic_create(t *testing.T) {
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

func TestPost_create(t *testing.T) {
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

func TestMode_update(t *testing.T) {
	mode := Mode{
		Id: 1,
	}

	b := struct {
		Name *string
		Pub  *bool
	}{
		Name: nil,
		Pub:  ref(true),
	}

	err := mode.update(b)
	if err != nil {
		t.Error(err)
	}
}

func TestTopic_update(t *testing.T) {
	topic := Topic{
		Id: 99,
	}

	b1 := struct {
		ModeId *int
		Title  *string
	}{
		ModeId: ref(1),
		Title:  ref("test"),
	}
	b2 := map[string]interface{}{
		"mode_id": ref(1),
		"title":   ref("test"),
	}

	_, _ = b1, b2
	err := topic.update(b1)
	if err != nil {
		t.Error(err)
	}
}

func TestPost_update(t *testing.T) {
	post := Post{
		TopicId: 1,
		Floor:   2,
	}

	b := Post{
		Content: "",
	}

	err := post.update(b)
	if err != nil {
		t.Error(err)
	}
}

func TestMode_delete(t *testing.T) {
	mode := Mode{
		Id: 1,
	}

	err := mode.delete()
	if err != nil {
		t.Error(err)
	}
}

func TestTopic_delete(t *testing.T) {
	topic := Topic{
		Id: 92,
	}

	err := topic.delete()
	if err != nil {
		t.Error(err)
	}
}

func TestPost_delete(t *testing.T) {
	post := Post{
		TopicId: 2,
		Floor:   1,
	}

	err := post.delete()
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

func TestTopic_stat(t *testing.T) {
	a := Topic{
		Id: 1,
	}

	err := a.stat("id")
	if err != nil {
		t.Error(err)
	}
}

func ref[T any](x T) *T {
	return &x
}
