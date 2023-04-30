package routes

import (
	"github.com/cli-ish/deezer-badge/internal/util"
	"html/template"
	"net/http"
	"time"
)

func (bs *BadgeServer) getAuth(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	if queryValues.Get("code") == "" {
		lifeTime := 30 * time.Minute
		cookie := util.GenerateSecureToken(64)
		sessKey := util.GenerateSecureToken(32)
		err := bs.Database.SetCookie(cookie, sessKey, lifeTime)
		if err != nil {
			http.Error(w, "redis down", 500)
			return
		}
		myCookie := &http.Cookie{Name: "session", Value: cookie, HttpOnly: true, Path: "/", Expires: time.Now().Add(lifeTime)}
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
		sessKey, err := bs.Database.GetCookie(cookie.Value)
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

		uidStr, err := bs.Database.CreateUser(userData.Id, token)
		if err != nil {
			http.Error(w, "redis down", 500)
			return
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
