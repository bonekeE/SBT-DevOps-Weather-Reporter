package weather

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

type MockRedisClient struct {
	storage map[string]string
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if val, ok := m.storage[key]; ok {
		return redis.NewStringResult(val, nil)
	}
	return redis.NewStringResult("", redis.Nil)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	m.storage[key] = value.(string)
	return redis.NewStatusResult("OK", nil)
}

func TestWeatherHandler(t *testing.T) {
	mockRedis := &MockRedisClient{storage: make(map[string]string)}
	redisClient = mockRedis

	mockResponse := WeatherResponse{
		City: "Moscow",
		Temp: "5.0°C",
		Desc: "clear sky",
	}
	mockData, _ := json.Marshal(mockResponse)
	mockRedis.Set(context.Background(), "Moscow", string(mockData), cacheExpiration)

	payload := `{"cities": ["Moscow"]}`
	req := httptest.NewRequest(http.MethodPost, "/weather", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	WeatherHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responses []WeatherResponse
	err := json.NewDecoder(res.Body).Decode(&responses)
	assert.NoError(t, err)

	assert.Len(t, responses, 2)
	assert.Equal(t, "Moscow", responses[0].City)
	assert.Equal(t, "5.0°C", responses[0].Temp)
	assert.Equal(t, "clear sky", responses[0].Desc)
	assert.Empty(t, responses[0].Error)
}
