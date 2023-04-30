package routes

import (
	"encoding/base64"
	"fmt"
	"github.com/cli-ish/deezer-badge/internal/models"
	"github.com/cli-ish/deezer-badge/internal/util"
	"html/template"
	"net/http"
	"strings"
	"time"
)

func (bs *BadgeServer) getBadge(w http.ResponseWriter, r *http.Request) {
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
	userId, err := bs.RedisClient.Get(bs.ctx, "uid:"+uidStr).Result()
	if err != nil {
		http.Error(w, "redis down", 500)
		return
	}
	accessToken, err := bs.RedisClient.Get(bs.ctx, "user:"+fmt.Sprint(userId)).Result()
	if err != nil {
		http.Error(w, "redis down", 500)
		return
	}

	historyResult, err := bs.DeezerApi.GetUserHistory(userId, accessToken, bs.RedisClient, bs.ctx)
	if err != nil {
		http.Error(w, "invalid history information response", 500)
		return
	}
	lastPlayedTrack := models.BasicHistoryInfo{}
	if len(historyResult.Data) > 0 {
		lastPlayedTrack = historyResult.Data[0]
	}
	if lastPlayedTrack.Album.CoverMedium != "" && strings.HasPrefix(lastPlayedTrack.Album.CoverMedium, "http") {
		data, err := util.GetPageContentCached(lastPlayedTrack.Album.CoverMedium, bs.RedisClient, bs.ctx, 30*time.Minute)
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
