package main

import (
	"fmt"
	"log"
	"net/http"
	"weather-service/weather"
)

func main() {
	// Инициализируем Redis
	weather.InitRedis()

	// Обработчик для пути /weather
	http.HandleFunc("/weather", weather.WeatherHandler)

	// Запуск сервера
	fmt.Println("Сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
