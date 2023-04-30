package routes

import (
	"fmt"
	"github.com/cli-ish/deezer-badge/internal/util"
	"github.com/google/uuid"
	"html/template"
	"net/http"
	"time"
)

func (bs *BadgeServer) getAuth(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	if queryValues.Get("code") == "" {
		lifetime := 30 * time.Minute
		cookie := util.GenerateSecureToken(64)
		sessKey := util.GenerateSecureToken(32)
		err := bs.RedisClient.Set(bs.ctx, "cookie:"+cookie, sessKey, lifetime).Err()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		myCookie := &http.Cookie{Name: "session", Value: cookie, HttpOnly: true, Path: "/", Expires: time.Now().Add(lifetime)}
		http.SetCookie(w, myCookie)
		http.Redirect(w, r, bs.DeezerApi.GenerateRedirectUrl(sessKey), 302)
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
		sessKey, err := bs.RedisClient.Get(bs.ctx, "cookie:"+cookie.Value).Result()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		if queryValues.Get("state") != sessKey {
			http.Error(w, "CSRF token does not match", 500)
			return
		}

		token, err := bs.DeezerApi.GetAccessToken(queryValues.Get("code"))
		if err != nil {
			http.Error(w, "access token could not be loaded", 500)
			return
		}

		userData, err := bs.DeezerApi.GetMe(token)
		if err != nil {
			http.Error(w, "me information parsing error", 500)
			return
		}

		err = bs.RedisClient.Set(bs.ctx, "user:"+fmt.Sprint(userData.Id), token, 0).Err()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		userUidKey := "user:" + fmt.Sprint(userData.Id) + ":uid"
		exist, err := bs.RedisClient.Exists(bs.ctx, userUidKey).Result()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		uidStr := ""
		if exist == 0 {
			uidStr = uuid.NewString()
			err = bs.RedisClient.MSet(bs.ctx, map[string]string{userUidKey: uidStr, "uid:" + uidStr: fmt.Sprint(userData.Id)}).Err()
			if err != nil {
				http.Error(w, "redis down", 500)
				return
			}
		} else {
			uidStr, err = bs.RedisClient.Get(bs.ctx, userUidKey).Result()
			if err != nil {
				http.Error(w, "redis down", 500)
				return
			}
		}
		tmpl := template.Must(template.ParseFS(badgeTemplate, "templates/auth_result.gohtml", "templates/head.gohtml", "templates/footer.gohtml"))
		w.Header().Set("Content-Type", "text/html")
		err = tmpl.Execute(w, struct {
			Title    string
			BadgeUrl string
		}{Title: "Your badge is ready!", BadgeUrl: "./badge/" + uidStr})
		if err != nil {
			http.Error(w, "unable to template index", 500)
			return
		}
		return
	} else {
		http.Error(w, "CSRF token does not match", 500)
		return
	}
}
