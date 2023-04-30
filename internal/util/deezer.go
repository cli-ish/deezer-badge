package util

import (
	"context"
	"encoding/json"
	"github.com/cli-ish/deezer-badge/internal/models"
	"github.com/redis/go-redis/v9"
	"net/url"
	"time"
)

type CustomDeezerApi struct {
	AppId     string
	AppSecret string
	ReturnUrl string
}

func (cda *CustomDeezerApi) GetMe(accessToken string) (models.BasicUserInfo, error) {
	meUrl := "https://api.deezer.com/user/me?"
	values := url.Values{}
	values.Set("access_token", accessToken)
	bodyBytes, err := GetPageContent(meUrl + values.Encode())
	data := models.BasicUserInfo{}
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func (cda *CustomDeezerApi) GetAccessToken(code string) (string, error) {
	tokenUrl := "https://connect.deezer.com/oauth/access_token.php?"
	values := url.Values{}
	values.Set("app_id", cda.AppId)
	values.Set("secret", cda.AppSecret)
	values.Set("code", code)
	bodyBytes, err := GetPageContent(tokenUrl + values.Encode())
	if err != nil {
		return "", err
	}
	responseQuery, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return "", err
	}
	token := responseQuery.Get("access_token")
	if token == "" {
		return "", err
	}
	return token, nil
}

func (cda *CustomDeezerApi) GetUserHistory(userId string, accessToken string, redisClient *redis.Client, ctx context.Context) (models.BasicWrapHistory, error) {
	historyUrl := "https://api.deezer.com/user/" + userId + "/history?"
	values := url.Values{}
	values.Set("access_token", accessToken)
	var historyResult models.BasicWrapHistory
	resultData, err := GetPageContentCached(historyUrl+values.Encode(), redisClient, ctx, 30*time.Second)
	if err != nil {
		return historyResult, err
	}
	err = json.Unmarshal(resultData, &historyResult)
	if err != nil {
		return historyResult, err
	}
	return historyResult, nil
}

func (cda *CustomDeezerApi) GenerateRedirectUrl(sessKey string) string {
	values := url.Values{}
	values.Set("app_id", cda.AppId)
	values.Set("redirect_uri", cda.ReturnUrl)
	values.Set("perms", "email,offline_access,listening_history")
	values.Set("state", sessKey)
	return "https://connect.deezer.com/oauth/auth.php?" + values.Encode()
}
