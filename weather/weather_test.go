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
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
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
	runner.Run(t, "TestWeatherHandler", func(t provider.T) {
		mockRedis := &MockRedisClient{storage: make(map[string]string)}
		redisClient = mockRedis

		t.WithNewStep("Set up mock Redis data", func(sCtx provider.StepCtx) {
			mockResponse := WeatherResponse{
				City: "Moscow",
				Temp: "5.0°C",
				Desc: "clear sky",
			}
			mockData, err := json.Marshal(mockResponse)
			sCtx.Assert().NoError(err, "Mock data marshalling should not fail")
			mockRedis.Set(context.Background(), "Moscow", string(mockData), cacheExpiration)
		})

		payload := `{"cities": ["Moscow"]}`
		var rec *httptest.ResponseRecorder

		t.WithNewStep("Create and send HTTP request", func(sCtx provider.StepCtx) {
			req := httptest.NewRequest(http.MethodPost, "/weather", bytes.NewBuffer([]byte(payload)))
			req.Header.Set("Content-Type", "application/json")
			rec = httptest.NewRecorder()
			WeatherHandler(rec, req)
		})

		t.WithNewStep("Verify HTTP response", func(sCtx provider.StepCtx) {
			res := rec.Result()
			defer res.Body.Close()
			sCtx.Assert().Equal(http.StatusOK, res.StatusCode, "Response status code should be 200 OK")

			var responses []WeatherResponse
			err := json.NewDecoder(res.Body).Decode(&responses)
			sCtx.Assert().NoError(err, "Response decoding should not fail")

			sCtx.Assert().Len(responses, 1, "Response should contain exactly one item")
			sCtx.Assert().Equal("Moscow", responses[0].City, "City name should match")
			sCtx.Assert().Equal("5.0°C", responses[0].Temp, "Temperature should match")
			sCtx.Assert().Equal("clear sky", responses[0].Desc, "Description should match")
			sCtx.Assert().Empty(responses[0].Error, "Error field should be empty")
		})
	})
}
