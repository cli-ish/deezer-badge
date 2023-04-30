package routes

import (
	"embed"
	"github.com/cli-ish/deezer-badge/internal/util"
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
	Port      string
	DeezerApi util.CustomDeezerApi
	Database  util.Database
}

func (bs *BadgeServer) Init() {
	bs.DeezerApi = util.CustomDeezerApi{
		AppId:     os.Getenv("APP_ID"),
		AppSecret: os.Getenv("APP_SECRET"),
		ReturnUrl: os.Getenv("RETURN_URL"),
	}

	bs.Database = util.Database{}
	bs.Database.Init()

	bs.Port = os.Getenv("HTTP_PORT")
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
