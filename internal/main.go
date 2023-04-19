package main

import (
	"context"
	"embed"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cli-ish/deezer-badge/internal/models"
	"github.com/cli-ish/deezer-badge/internal/util"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

//go:embed templates/badge.gohtml
var badgeTemplate embed.FS

//go:embed terms.txt
var termsFile embed.FS

var cookieExp = regexp.MustCompile(`^[a-zA-Z0-9]{128}$`)
var uidExp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

var rdb *redis.Client
var ctx = context.Background()
var appId = ""
var appSecret = ""
var returnUrl = ""

func main() {
	appId = os.Getenv("APP_ID")
	appSecret = os.Getenv("APP_SECRET")
	returnUrl = os.Getenv("RETURN_URL")
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/auth", getAuth)
	mux.HandleFunc("/badge/", getBadge)
	mux.HandleFunc("/terms", getTerms)
	log.Println("Starting server on port :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func getAuth(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	if queryValues.Get("code") == "" {
		lifetime := 30 * time.Minute
		cookie := util.GenerateSecureToken(64)
		sessKey := util.GenerateSecureToken(32)
		err := rdb.Set(ctx, "cookie:"+cookie, sessKey, lifetime).Err()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		values := url.Values{}
		values.Set("app_id", appId)
		values.Set("redirect_uri", returnUrl)
		values.Set("perms", "email,offline_access,listening_history")
		values.Set("state", sessKey)
		myCookie := &http.Cookie{Name: "session", Value: cookie, HttpOnly: true, Path: "/", Expires: time.Now().Add(lifetime)}
		http.SetCookie(w, myCookie)
		http.Redirect(w, r, "https://connect.deezer.com/oauth/auth.php?"+values.Encode(), 302)
		return
	}
	if queryValues.Get("state") != "" {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "session cookie not found", 500)
			return
		}
		if !cookieExp.MatchString(cookie.Value) {
			http.Error(w, "invalid session cookie", 500)
			return
		}
		sessKey, err := rdb.Get(ctx, "cookie:"+cookie.Value).Result()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		if queryValues.Get("state") != sessKey {
			http.Error(w, "CSRF token does not match", 500)
			return
		}
		tokenUrl := "https://connect.deezer.com/oauth/access_token.php?"
		values := url.Values{}
		values.Set("app_id", appId)
		values.Set("secret", appSecret)
		values.Set("code", queryValues.Get("code"))
		bodyBytes, err := util.GetPageContent(tokenUrl + values.Encode())
		if err != nil {
			http.Error(w, "token fetching failed", 500)
			return
		}
		responseQuery, err := url.ParseQuery(string(bodyBytes))
		if err != nil {
			http.Error(w, "result parser failed", 500)
			return
		}
		token := responseQuery.Get("access_token")
		if token == "" {
			log.Println("invalid access token")
			return
		}

		meUrl := "https://api.deezer.com/user/me?"
		values = url.Values{}
		values.Set("access_token", token)
		bodyBytes, err = util.GetPageContent(meUrl + values.Encode())
		if err != nil {
			http.Error(w, "me information fetching failed", 500)
			return
		}

		data := models.BasicUserInfo{}
		err = json.Unmarshal(bodyBytes, &data)
		if err != nil {
			http.Error(w, "me information parsing error", 500)
			return
		}

		err = rdb.Set(ctx, "user:"+fmt.Sprint(data.Id), token, 0).Err()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		userUidKey := "user:" + fmt.Sprint(data.Id) + ":uid"
		exist, err := rdb.Exists(ctx, userUidKey).Result()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		uidStr := ""
		if exist == 0 {
			uidStr = uuid.NewString()
			err = rdb.MSet(ctx, map[string]string{userUidKey: uidStr, "uid:" + uidStr: fmt.Sprint(data.Id)}).Err()
			if err != nil {
				http.Error(w, "redis down", 500)
				return
			}
		} else {
			uidStr, err = rdb.Get(ctx, userUidKey).Result()
			if err != nil {
				http.Error(w, "redis down", 500)
				return
			}
		}
		http.Redirect(w, r, "badge/"+uidStr, 301)
		return
	} else {
		http.Error(w, "CSRF token does not match", 500)
		return
	}
}

func getBadge(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	uidStr := parts[len(parts)-1]
	if uidStr == "" {
		http.Error(w, "invalid uid", 500)
		return
	}
	if !uidExp.MatchString(uidStr) {
		http.Error(w, "invalid uid", 500)
		return
	}
	userId, err := rdb.Get(ctx, "uid:"+uidStr).Result()
	if err != nil {
		http.Error(w, "redis down", 500)
		return
	}
	accessToken, err := rdb.Get(ctx, "user:"+fmt.Sprint(userId)).Result()
	if err != nil {
		http.Error(w, "redis down", 500)
		return
	}

	historyUrl := "https://api.deezer.com/user/" + fmt.Sprint(userId) + "/history?"
	values := url.Values{}
	values.Set("access_token", accessToken)

	resultData, err := util.GetPageContentCached(historyUrl+values.Encode(), rdb, ctx, 30*time.Second)
	if err != nil {
		http.Error(w, "could not be requested", 500)
		return
	}

	var historyResult models.BasicWrapHistory
	err = json.Unmarshal(resultData, &historyResult)
	if err != nil {
		http.Error(w, "invalid history information response", 500)
		return
	}
	lastPlayedTrack := models.BasicHistoryInfo{}
	if len(historyResult.Data) > 0 {
		lastPlayedTrack = historyResult.Data[0]
	}
	if lastPlayedTrack.Album.CoverMedium != "" && strings.HasPrefix(lastPlayedTrack.Album.CoverMedium, "http") {
		data, err := util.GetPageContentCached(lastPlayedTrack.Album.CoverMedium, rdb, ctx, 30*time.Minute)
		if err != nil {
			http.Error(w, "could not be requested", 500)
			return
		}
		lastPlayedTrack.BasicImage = base64.StdEncoding.EncodeToString(data)
	}
	tmpl := template.Must(template.ParseFS(badgeTemplate, "templates/badge.gohtml"))
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache")
	err = tmpl.Execute(w, lastPlayedTrack)
	if err != nil {
		http.Error(w, "unable to template badge", 500)
	}
}

func getTerms(w http.ResponseWriter, _ *http.Request) {
	terms, err := termsFile.ReadFile("terms.txt")
	if err != nil {
		http.Error(w, "unable to serve terms", 500)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	_, err = w.Write(terms)
	if err != nil {
		http.Error(w, "unable to serve terms", 500)
		return
	}
}
