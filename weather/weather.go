package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	apiKey          = "07a8a4a26ad711fe07ba1a21e24fe9d6"
	redisDB         = 0
	cacheExpiration = 20 * time.Second
)

type RedisClientInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

var redisClient RedisClientInterface

type WeatherRequest struct {
	Cities []string `json:"cities"`
}

type WeatherResponse struct {
	City  string `json:"city"`
	Temp  string `json:"temperature"`
	Desc  string `json:"description"`
	Error string `json:"error,omitempty"`
}

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":6379",
		DB:   redisDB,
	})
}

func WeatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается, используйте POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req WeatherRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Ошибка разбора JSON", http.StatusBadRequest)
		return
	}

	var responses []WeatherResponse
	for _, city := range req.Cities {
		response, err := getWeather(city)
		if err != nil {
			responses = append(responses, WeatherResponse{City: city, Error: err.Error()})
			continue
		}
		responses = append(responses, *response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func getWeather(city string) (*WeatherResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cached, err := redisClient.Get(ctx, city).Result()
	if err == nil {
		var response WeatherResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			fmt.Println("Данные из кэша для:", city)
			return &response, nil
		}
	}

	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s", city, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API для города %s: %s", city, resp.Status)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа API")
	}

	main, ok := data["main"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("не найдено поле main в ответе API")
	}
	weatherList, ok := data["weather"].([]interface{})
	if !ok || len(weatherList) == 0 {
		return nil, fmt.Errorf("не найдено поле weather в ответе API")
	}
	weather, ok := weatherList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ошибка в формате weather")
	}

	response := &WeatherResponse{
		City: city,
		Temp: fmt.Sprintf("%.1f°C", main["temp"].(float64)),
		Desc: weather["description"].(string),
	}

	dataToCache, _ := json.Marshal(response)
	redisClient.Set(ctx, city, dataToCache, cacheExpiration)

	return response, nil
}
