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
	colorBarrier := byte(66)
	result := [3]byte{colorBarrier, colorBarrier, colorBarrier}
	textB := []byte(text)
	textLen := len(textB)
	for i := 0; i < textLen; i += 3 {
		c1, c2 := byte(0), byte(0)
		if i+1 < textLen {
			c1 = textB[i+1]
		}
		if i+2 < textLen {
			c2 = textB[i+2]
		}
		result = [3]byte{result[0] ^ textB[i], result[1] ^ c1, result[2] ^ c2}
	}
	limit := int(colorBarrier) * 4
	if int(result[0])+int(result[1])+int(result[2]) < limit {
		result[0] += colorBarrier * 2
	}
	if int(result[0])+int(result[1])+int(result[2]) < limit {
		result[1] += colorBarrier * 2
	}
	if int(result[0])+int(result[1])+int(result[2]) < limit {
		result[2] += colorBarrier * 2
	}
	return fmt.Sprintf("%x", result)
}
