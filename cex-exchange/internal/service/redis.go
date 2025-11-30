package service

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/v2/frame/g"
)

// Redis Service
type sRedis struct {
	Client *redis.Client
}

var Redis = &sRedis{}

// Init 初始化 Redis 客户端
func (s *sRedis) Init() {
	s.Client = redis.NewClient(&redis.Options{
		Addr:         "194.164.194.118:9502",
		Password:     "pass123editmelol",
		DB:           0,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := s.Client.Ping(ctx).Result(); err != nil {
		g.Log().Errorf(ctx, "❌ Redis connection failed (v8): %v", err)
	} else {
		g.Log().Info(ctx, "✅ Redis connected successfully (v8)!")
	}
}

// Get retrieves a value from Redis
func (s *sRedis) Get(ctx context.Context, key string) (string, error) {
	return s.Client.Get(ctx, key).Result()
}

// Set stores a value in Redis
func (s *sRedis) Set(ctx context.Context, key string, value interface{}) error {
	return s.Client.Set(ctx, key, value, 0).Err()
}

// SetEX stores a value in Redis with expiration time (in seconds)
func (s *sRedis) SetEX(ctx context.Context, key string, value interface{}, seconds int64) error {
	return s.Client.Set(ctx, key, value, time.Duration(seconds)*time.Second).Err()
}

// Del deletes one or more keys from Redis
func (s *sRedis) Del(ctx context.Context, keys ...string) error {
	return s.Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in Redis
func (s *sRedis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.Client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire sets a timeout on a key
func (s *sRedis) Expire(ctx context.Context, key string, seconds int64) error {
	return s.Client.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
}

// HSet stores a value in a hash
func (s *sRedis) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return s.Client.HSet(ctx, key, field, value).Err()
}

// HGet retrieves a value from a hash
func (s *sRedis) HGet(ctx context.Context, key string, field string) (string, error) {
	return s.Client.HGet(ctx, key, field).Result()
}

// HGetAll retrieves all fields and values from a hash
func (s *sRedis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return s.Client.HGetAll(ctx, key).Result()
}

// HDel deletes one or more fields from a hash
func (s *sRedis) HDel(ctx context.Context, key string, fields ...string) error {
	return s.Client.HDel(ctx, key, fields...).Err()
}

// Incr increments the integer value of a key by one
func (s *sRedis) Incr(ctx context.Context, key string) (int64, error) {
	return s.Client.Incr(ctx, key).Result()
}

// IncrBy increments the integer value of a key by the given amount
func (s *sRedis) IncrBy(ctx context.Context, key string, increment int64) (int64, error) {
	return s.Client.IncrBy(ctx, key, increment).Result()
}

// Decr decrements the integer value of a key by one
func (s *sRedis) Decr(ctx context.Context, key string) (int64, error) {
	return s.Client.Decr(ctx, key).Result()
}

// Do executes a Redis command directly
func (s *sRedis) Do(ctx context.Context, command string, args ...interface{}) (interface{}, error) {
	cmdArgs := make([]interface{}, 0, 1+len(args))
	cmdArgs = append(cmdArgs, command)
	cmdArgs = append(cmdArgs, args...)
	return s.Client.Do(ctx, cmdArgs...).Result()
}
