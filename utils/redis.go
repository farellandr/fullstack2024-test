package utils

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

func InitRedis() *RedisClient {
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisAddr := redisHost + ":" + redisPort

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Println("Connected to Redis")
	}

	return &RedisClient{
		Client: client,
		Ctx:    ctx,
	}
}

func (r *RedisClient) SetClientData(slug string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.Client.Set(r.Ctx, "client:"+slug, jsonData, 0).Err()
}

func (r *RedisClient) GetClientData(slug string) (string, error) {
	return r.Client.Get(r.Ctx, "client:"+slug).Result()
}

func (r *RedisClient) DeleteClientData(slug string) error {
	return r.Client.Del(r.Ctx, "client:"+slug).Err()
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
