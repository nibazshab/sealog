package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type config struct {
	port     string
	database string
	log      string
	w        io.Writer
}

func main() {
	cfg := initializeApplication()
	initializeDbDrive(cfg)
	initializeLogDrive(cfg)
	run(cfg)
}

func run(cfg *config) {
	initializeAdminUser()
	initializeJwtSecret()

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	gin.DefaultWriter = cfg.w

	r := gin.Default()
	r.Use(corsMiddleware())

	// r.GET()

	srv := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: r,
	}

	log.Println("Listening on " + cfg.port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln("error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	closeDb()
}

func initializeApplication() *config {
	repw := flag.Bool("reset-password", false, "reset admin password")
	port := flag.Int("p", 8080, "server port")

	flag.Parse()

	if *repw {
		pw, err := resetAdminPassword()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("new password:", pw)
		os.Exit(0)
	}

	if *port < 1 || *port > 65535 {
		fmt.Println("error:", "invalid port")
	}

	ex, _ := os.Executable()
	dir := filepath.Join(filepath.Dir(ex), "data")
	fi, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = os.MkdirAll(dir, 0o755)
			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	} else {
		if !fi.IsDir() {
			fmt.Println("error:", dir+" not a directory")
			os.Exit(1)
		}
	}

	return &config{
		port:     strconv.Itoa(*port),
		database: filepath.Join(dir, "data.db"),
		log:      filepath.Join(dir, "log.log"),
	}
}

func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}
