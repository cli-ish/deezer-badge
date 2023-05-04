package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

func GenerateColorCode(text string) string {
	result := [3]byte{0, 0, 0}
	textB := []byte(text)
	textLen := len(textB)
	for i := 0; i < textLen; i += 3 {
		c0, c1, c2 := textB[i], byte(0), byte(0)
		if i+1 < textLen {
			c1 = textB[i+1]
		}
		if i+2 < textLen {
			c2 = textB[i+2]
		}
		result = [3]byte{result[0] ^ c0, result[1] ^ c1, result[2] ^ c2}
	}
	return fmt.Sprintf("%x", result)
}
