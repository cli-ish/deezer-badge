package routes

import (
	"encoding/json"
	"fmt"
	"github.com/cli-ish/deezer-badge/internal/models"
	"github.com/cli-ish/deezer-badge/internal/util"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"net/url"
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
		values := url.Values{}
		values.Set("app_id", bs.AppId)
		values.Set("redirect_uri", bs.ReturnUrl)
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
		sessKey, err := bs.RedisClient.Get(bs.ctx, "cookie:"+cookie.Value).Result()
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
		values.Set("app_id", bs.AppId)
		values.Set("secret", bs.AppSecret)
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

		err = bs.RedisClient.Set(bs.ctx, "user:"+fmt.Sprint(data.Id), token, 0).Err()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		userUidKey := "user:" + fmt.Sprint(data.Id) + ":uid"
		exist, err := bs.RedisClient.Exists(bs.ctx, userUidKey).Result()
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		uidStr := ""
		if exist == 0 {
			uidStr = uuid.NewString()
			err = bs.RedisClient.MSet(bs.ctx, map[string]string{userUidKey: uidStr, "uid:" + uidStr: fmt.Sprint(data.Id)}).Err()
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
			Title string
		}{"Your badge is ready!"})
		if err != nil {
			http.Error(w, "unable to template index", 500)
		}
		return
	} else {
		http.Error(w, "CSRF token does not match", 500)
		return
	}
}
