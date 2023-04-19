package util

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"time"
)

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func GetPageContent(url string) ([]byte, error) {
	result, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(result.Body)

	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

func GetPageContentCached(url string, rdb *redis.Client, ctx context.Context, cacheLength time.Duration) ([]byte, error) {
	var data []byte
	key := "url:" + url
	exist, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return []byte{}, err
	}
	if exist == 0 {
		data, err = GetPageContent(url)
		if err != nil {
			return []byte{}, err
		}
		_, err = rdb.Set(ctx, key, base64.StdEncoding.EncodeToString(data), cacheLength).Result()
		if err != nil {
			return []byte{}, err
		}
	} else {
		base64data, err := rdb.Get(ctx, key).Result()
		if err != nil {
			return []byte{}, err
		}
		data, err = base64.StdEncoding.DecodeString(base64data)
		if err != nil {
			return []byte{}, err
		}
	}

	return data, nil
}
