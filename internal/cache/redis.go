package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Ctx = context.Background()

const (
	FeatureTTL  = 5 * time.Minute
	AllFeatures = "features:all"
)

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379")),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	for i := 0; i < 10; i++ {
		_, err := RDB.Ping(Ctx).Result()
		if err == nil {
			break
		}
		log.Printf("⏳ Redis chưa sẵn sàng, thử lại sau 2s... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	log.Println("✅ Kết nối Redis thành công")
}

func SetJSON(key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RDB.Set(Ctx, key, b, ttl).Err()
}

func GetJSON(key string, dest any) error {
	val, err := RDB.Get(Ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func Delete(keys ...string) error {
	return RDB.Del(Ctx, keys...).Err()
}

func FeatureKey(name string) string {
	return fmt.Sprintf("feature:%s", name)
}

func RouteKey(method, path string) string {
	return fmt.Sprintf("route:%s:%s", method, path)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
