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

const defaultPort = 8080

type config struct {
	port   string
	rootfs string
	w      io.Writer
}

func main() {
	cfg := initializeApplication()
	argsExecute(cfg)
	initializeDbDrive(cfg)
	initializeLogDrive(cfg)
	serverRun(cfg)
}

func initializeApplication() *config {
	ex, err := os.Executable()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	dir := filepath.Join(filepath.Dir(ex), "data")

	info, err := os.Stat(dir)
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
		if !info.IsDir() {
			fmt.Println("error:", dir+" not a directory")
			os.Exit(1)
		}
	}

	return &config{
		port:   strconv.Itoa(defaultPort),
		rootfs: dir,
	}
}

func argsExecute(cfg *config) {
	if len(os.Args) < 2 {
		fmt.Println(man)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "server":
		args := flag.NewFlagSet("server", flag.ExitOnError)
		var port int

		args.IntVar(&port, "p", defaultPort, "server port")

		err := args.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		if port < 1 || port > 65535 {
			fmt.Println("error:", "invalid port")
			os.Exit(1)
		}

		cfg.port = strconv.Itoa(port)

	case "reset-password":
		initializeDbDrive(cfg)
		rawPassword, err := resetAdminPassword()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("new password:", rawPassword)
		closeDb()
		os.Exit(0)

	case "-h", "--help":
		fmt.Println(man)
		os.Exit(0)

	default:
		fmt.Println("unknown", os.Args[1])
		os.Exit(0)
	}
}

func serverRun(cfg *config) {
	initializeAdminUser()
	initializeJwtSecret()

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	gin.DefaultWriter = cfg.w

	r := gin.Default()
	initializeRouter(r)

	srv := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: r,
	}

	go func() {
		log.Println("Listening on " + cfg.port)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln("error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	closeDb()
}

func initializeRouter(r *gin.Engine) {
	r.Use(corsMiddleware())
	r.Use(authMiddleware())

	p := r.Group("/p")
	p.GET("/", getDiscussions)
	p.GET("/:pid", getDiscussion)

	t := r.Group("/t")
	t.GET("/", getCategories)
	t.GET("/:tid", getDiscussionsByCategory)

	u := r.Group("/u")
	u.GET("/me", getUserInformation)

	api := r.Group("/api")
	api.Use(protectMiddleware())

	user := api.Group("/user")
	user.POST("/login", userLogin)
	user.POST("/update", userChangePassword)

	category := api.Group("/category")
	category.POST("/create", createCategory)
	category.POST("/update", updateCategory)
	category.POST("/delete", deleteCategory)

	discussion := api.Group("/discussion")
	discussion.POST("/create", createDiscussion)
	discussion.POST("/update", updateDiscussion)
	discussion.POST("/delete", deleteDiscussion)

	comment := api.Group("/comment")
	comment.POST("/create", createComment)
	comment.POST("/update", updateComment)
	comment.POST("/delete", deleteComment)
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

const man = `available command:
  server          start httpserver
  reset-password  reset admin password`
