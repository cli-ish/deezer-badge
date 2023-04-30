package util

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

type Database struct {
	redisClient  *redis.Client
	ctx          context.Context
	prefixCookie string
	prefixCache  string
	prefixUid    string
	prefixUser   string
}

func (d *Database) Init() {
	d.redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})
	d.ctx = context.Background()
	d.prefixCookie = "cookie:"
	d.prefixCache = "url:"
	d.prefixUid = "uid:"
	d.prefixUser = "user:"
}

func (d *Database) SetCookie(cookie string, sessKey string, lifeTime time.Duration) error {
	return d.redisClient.Set(d.ctx, d.prefixCookie+cookie, sessKey, lifeTime).Err()
}

func (d *Database) GetCookie(cookie string) (string, error) {
	return d.redisClient.Get(d.ctx, d.prefixCookie+cookie).Result()
}

func (d *Database) CreateUser(userId int64, accessToken string) (string, error) {
	err := d.redisClient.Set(d.ctx, d.prefixUser+fmt.Sprint(userId), accessToken, 0).Err()
	if err != nil {
		return "", err
	}
	userUidKey := d.prefixUser + fmt.Sprint(userId) + ":uid"
	exist, err := d.redisClient.Exists(d.ctx, userUidKey).Result()
	if err != nil {
		return "", err
	}
	uidStr := ""
	if exist == 0 {
		uidStr = uuid.NewString()
		err = d.redisClient.MSet(d.ctx, map[string]string{userUidKey: uidStr, d.prefixUid + uidStr: fmt.Sprint(userId)}).Err()
		if err != nil {
			return "", err
		}
	} else {
		uidStr, err = d.redisClient.Get(d.ctx, userUidKey).Result()
		if err != nil {
			return "", err
		}
	}
	return uidStr, nil
}

func (d *Database) GetUserIdAndAccessToken(uid string) (string, string, error) {
	userId, err := d.redisClient.Get(d.ctx, d.prefixUid+uid).Result()
	if err != nil {
		return "", "", err
	}
	accessToken, err := d.redisClient.Get(d.ctx, d.prefixUser+userId).Result()
	if err != nil {
		return "", "", err
	}
	return userId, accessToken, nil
}

func (d *Database) ExistInUrlCache(url string) (bool, error) {
	exist, err := d.redisClient.Exists(d.ctx, d.prefixCache+url).Result()
	if err != nil {
		return false, err
	}
	return exist != 0, nil
}

func (d *Database) SetUrlCache(url string, content []byte, cacheLength time.Duration) error {
	_, err := d.redisClient.Set(d.ctx, d.prefixCache+url, base64.StdEncoding.EncodeToString(content), cacheLength).Result()
	return err
}

func (d *Database) GetUrlCache(url string) ([]byte, error) {
	base64data, err := d.redisClient.Get(d.ctx, d.prefixCache+url).Result()
	if err != nil {
		return []byte{}, err
	}
	return base64.StdEncoding.DecodeString(base64data)
}
