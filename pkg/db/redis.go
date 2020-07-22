package db

import (
	"goshop/service-product/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var (
	Redis *redis.Client
)

//连接redis
func GetRedisClient() (*redis.Client, error) {
	redisOption := &redis.Options{
		Addr: utils.C.Redis.Host,
		DB:   utils.C.Redis.Database,
	}

	if gin.Mode() == "prod" {
		redisOption.Password = utils.C.Redis.Password
	}

	client := redis.NewClient(redisOption)
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	Redis = client

	return client, nil
}
