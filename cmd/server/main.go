// Command server запускает HTTP-сервер ТИШКИ.
//
// Точка входа умышленно короткая: она только собирает зависимости
// (конфиг -> стор -> движок квиза -> HTTP-сервер) и запускает
// http.ListenAndServe. Вся логика живёт в internal/*, чтобы её можно
// было использовать и тестировать отдельно от процесса запуска сервера.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/tishka/tishka/internal/config"
	"github.com/tishka/tishka/internal/httpserver"
	"github.com/tishka/tishka/internal/quiz"
	"github.com/tishka/tishka/internal/store"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := config.Load()

	testStore := store.NewJSONStore(cfg.TestsDir)
	if err := testStore.Load(); err != nil {
		log.Error("не удалось загрузить тесты", "error", err)
		os.Exit(1)
	}

	engine := quiz.NewEngine()

	srv, err := httpserver.New(cfg, testStore, engine, log)
	if err != nil {
		log.Error("не удалось инициализировать сервер", "error", err)
		os.Exit(1)
	}

	log.Info("запускаемся", "port", cfg.Port, "tests_dir", cfg.TestsDir)

	if err := http.ListenAndServe(":"+cfg.Port, srv.Handler()); err != nil {
		log.Error("сервер остановился с ошибкой", "error", err)
		os.Exit(1)
	}
}
