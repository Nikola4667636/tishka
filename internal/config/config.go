// Package config собирает настройки приложения из переменных окружения.
// Все настройки в одном месте с явными дефолтами — чтобы запустить сервер
// локально не нужно было ничего настраивать, а для деплоя достаточно
// переопределить несколько ENV-переменных.
package config

import "os"

// Config — все настройки, которые нужны серверу для старта.
type Config struct {
	// Port — порт, на котором слушает HTTP-сервер.
	Port string
	// TestsDir — папка с JSON-файлами тестов (см. data/tests).
	TestsDir string
	// TemplatesDir — папка с HTML-шаблонами.
	TemplatesDir string
	// StaticDir — папка со статикой (css/js/картинки), которая отдаётся как есть.
	StaticDir string
	// BrandName — название бренда, подставляется в шаблоны (шапка, футер, <title>).
	BrandName string
}

// Load читает конфиг из переменных окружения, подставляя разумные дефолты
// для локального запуска "из коробки" — go run ./cmd/server без единой
// переменной окружения должен просто заработать.
func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8080"),
		TestsDir:     getEnv("TESTS_DIR", "data/tests"),
		TemplatesDir: getEnv("TEMPLATES_DIR", "web/templates"),
		StaticDir:    getEnv("STATIC_DIR", "web/static"),
		BrandName:    getEnv("BRAND_NAME", "ПРИЗМА"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
