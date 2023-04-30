package routes

import (
	"context"
	"embed"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"regexp"
)

//go:embed templates
var badgeTemplate embed.FS

//go:embed styles/simple.min.*
var stylesFiles embed.FS

var cookieExp = regexp.MustCompile(`^[a-zA-Z0-9]{128}$`)
var uidExp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type BadgeServer struct {
	Port        string
	AppId       string
	AppSecret   string
	ReturnUrl   string
	RedisClient *redis.Client
	ctx         context.Context
}

func (bs *BadgeServer) Init() {
	bs.AppId = os.Getenv("APP_ID")
	bs.AppSecret = os.Getenv("APP_SECRET")
	bs.ReturnUrl = os.Getenv("RETURN_URL")
	bs.RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})
	bs.Port = os.Getenv("HTTP_PORT")
	bs.ctx = context.Background()
}

func (bs *BadgeServer) Start() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", bs.getIndex)
	mux.HandleFunc("/auth", bs.getAuth)
	mux.HandleFunc("/badge/", bs.getBadge)
	mux.HandleFunc("/terms", bs.getTerms)
	mux.HandleFunc("/styles/", http.FileServer(http.FS(stylesFiles)).ServeHTTP)
	log.Println("Starting server on port :" + bs.Port)
	err := http.ListenAndServe(":"+bs.Port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
