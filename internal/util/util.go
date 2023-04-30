package util

import (
	"crypto/rand"
	"encoding/hex"
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

func GetPageContentCached(url string, database *Database, cacheLength time.Duration) ([]byte, error) {
	var data []byte
	exist, err := database.ExistInUrlCache(url)
	if !exist {
		data, err = GetPageContent(url)
		if err != nil {
			return []byte{}, err
		}
		err = database.SetUrlCache(url, data, cacheLength)
		if err != nil {
			return []byte{}, err
		}
	} else {
		data, err = database.GetUrlCache(url)
		if err != nil {
			return []byte{}, err
		}
	}
	return data, nil
}
