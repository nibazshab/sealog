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
	debug  bool
}

func main() {
	cfg := initializeApplication()
	commandExecute(cfg)
	initializeLogDrive(cfg)
	initializeDbDrive(cfg)
	initializeSrvDrive(cfg)
	initializeAuth()
	initializeHmac()
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

func commandExecute(cfg *config) {
	if len(os.Args) < 2 {
		fmt.Print(man)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "server":
		args := flag.NewFlagSet("server", flag.ExitOnError)
		var port int
		var debug bool

		args.IntVar(&port, "p", defaultPort, "server port")
		args.BoolVar(&debug, "debug", false, "debug mode")

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
		cfg.debug = debug

	case "reset-password":
		initializeDbDrive(cfg)
		fmt.Println("new password:", generatePassword())
		closeDb()
		os.Exit(0)

	case "-h", "--help":
		fmt.Print(man)
		os.Exit(0)

	default:
		fmt.Println("unknown", os.Args[1])
		os.Exit(1)
	}
}

func serverRun(cfg *config) {
	r := gin.Default()
	initializeRouter(r)

	srv := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: r,
	}

	go func() {
		fmt.Println("Listening on " + cfg.port)

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

	api := r.Group("/api")
	api.Use(authMiddleware())

	api.GET("/av", getTopics)
	api.GET("/cv", getModes)
	api.GET("/av/:aid", getTopicAndPosts)
	api.GET("/cv/:cid", getTopicsByMode)
	api.GET("/uid", getAuthUid)
	api.POST("/auth/login", loginAuth)

	api.Use(protectMiddleware())

	auth := api.Group("/auth")
	auth.POST("/change", changeAuth)

	cv := api.Group("/cv")
	cv.POST("/create", createMode)
	cv.POST("/update", updateMode)
	cv.POST("/delete", deleteMode)

	av := api.Group("/av")
	av.POST("/create", createTopic)
	av.POST("/update", updateTopic)
	av.POST("/delete", deleteTopic)

	fl := api.Group("/fl")
	fl.POST("/create", createPost)
	fl.POST("/update", updatePost)
	fl.POST("/delete", deletePost)

	s := r.Group("/")
	s.Use(cacheMiddleware())
	static(s)

	h := r.Group("/")
	h.Use(authMiddleware())
	renderHtml(h)
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

const man = `% command:
  server          start httpserver (use 'server -h' view help)
  reset-password  reset admin password
`
