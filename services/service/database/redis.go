package database

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"time"
)

func KeyMessageAckIndex(account string) string {
	return fmt.Sprintf("chat:ack:%s", account)
}

func InitRedis(addr string, pass string) (*redis.Client, error) {
	redisDB := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	})
	_, err := redisDB.Ping().Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return redisDB, nil
}

func InitFailoverRedis(masterName string, sentinelAddrs []string, password string, timeout time.Duration) (*redis.Client, error) {
	redisDB := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
		DialTimeout:   time.Second * 5,
		ReadTimeout:   timeout,
		WriteTimeout:  timeout,
	})
	_, err := redisDB.Ping().Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return redisDB, nil
}
