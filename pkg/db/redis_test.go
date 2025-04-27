package db

import (
	"context"
	"finly-backend/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func createTestRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func TestNewRedisDB_Success(t *testing.T) {
	ctx := context.Background()

	cfg := &config.Config{
		RedisHost:     "localhost",
		RedisPort:     "6379",
		RedisPassword: "",
		RedisDB:       0,
	}

	client, err := NewRedisDB(ctx, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	pong, err := client.Ping(ctx).Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", pong)
}

func TestNewRedisDB_Failure(t *testing.T) {
	ctx := context.Background()

	cfg := &config.Config{
		RedisHost:     "localhost",
		RedisPort:     "12345",
		RedisPassword: "",
		RedisDB:       0,
	}

	start := time.Now()
	client, err := NewRedisDB(ctx, cfg)
	elapsed := time.Since(start)

	assert.Nil(t, client)
	assert.Error(t, err)

	assert.GreaterOrEqual(t, int(elapsed.Seconds()), 10)
}

func TestSetAndGetFromCache(t *testing.T) {
	ctx := context.Background()
	client := createTestRedisClient()
	defer client.FlushDB(ctx)

	key := "test_key"
	expected := testStruct{Name: "Alice", Age: 30}

	err := SetToCache(ctx, client, key, expected, time.Minute)
	assert.NoError(t, err)

	var actual testStruct
	found, err := GetFromCache(ctx, client, key, &actual)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, expected, actual)
}

func TestWithCache_CacheMissAndSet(t *testing.T) {
	ctx := context.Background()
	client := createTestRedisClient()
	defer client.FlushDB(ctx)

	key := "new_key"
	expected := testStruct{Name: "Bob", Age: 25}

	result, err := WithCache(ctx, client, key, time.Minute, func() (testStruct, error) {
		return expected, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	var cached testStruct
	found, err := GetFromCache(ctx, client, key, &cached)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, expected, cached)
}

func TestWithCache_CacheHit(t *testing.T) {
	ctx := context.Background()
	client := createTestRedisClient()
	defer client.FlushDB(ctx)

	key := "existing_key"
	existing := testStruct{Name: "Charlie", Age: 40}

	err := SetToCache(ctx, client, key, existing, time.Minute)
	assert.NoError(t, err)

	fetchCalled := false
	result, err := WithCache(ctx, client, key, time.Minute, func() (testStruct, error) {
		fetchCalled = true
		return testStruct{}, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, existing, result)
	assert.False(t, fetchCalled)
}

func TestGetFromCache_NotFound(t *testing.T) {
	ctx := context.Background()
	client := createTestRedisClient()
	defer client.FlushDB(ctx)

	var result testStruct
	found, err := GetFromCache(ctx, client, "non_existing_key", &result)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestGetFromCache_InvalidData(t *testing.T) {
	ctx := context.Background()
	client := createTestRedisClient()
	defer client.FlushDB(ctx)

	key := "invalid_json"
	err := client.Set(ctx, key, "invalid_json", time.Minute).Err()
	assert.NoError(t, err)

	var result testStruct
	found, err := GetFromCache(ctx, client, key, &result)
	assert.Error(t, err)
	assert.False(t, found)
}
