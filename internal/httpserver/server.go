// Package httpserver — весь HTTP-слой приложения: маршрутизация,
// шаблоны страниц и JSON API. Слой намеренно "тонкий": он не содержит
// бизнес-логики подбора результата (это internal/quiz) и не знает,
// откуда берутся тесты (это internal/store) — только разбирает запросы,
// вызывает нужные зависимости и форматирует ответ.
package httpserver

import (
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/tishka/tishka/internal/config"
	"github.com/tishka/tishka/internal/quiz"
	"github.com/tishka/tishka/internal/store"
)

// Server держит все зависимости, нужные HTTP-хендлерам.
type Server struct {
	cfg       config.Config
	store     store.Store
	engine    quiz.Engine
	templates map[string]*template.Template
	log       *slog.Logger
}

// pageTemplates перечисляет пары (имя шаблона -> файл с его контентом).
// Чтобы добавить новую страницу: создать web/templates/имя.html с блоком
// {{define "content"}}...{{end}} и одной строкой добавить его сюда.
var pageTemplates = map[string]string{
	"index":  "index.html",
	"tests":  "tests.html",
	"quiz":   "quiz.html",
	"prints": "prints.html",
	"print":  "print.html",
}

// New собирает Server: загружает и парсит HTML-шаблоны из cfg.TemplatesDir.
// Возвращает ошибку, если хотя бы один шаблон не парсится — сервер не
// должен стартовать с битыми шаблонами.
func New(cfg config.Config, st store.Store, engine quiz.Engine, log *slog.Logger) (*Server, error) {
	layoutPath := filepath.Join(cfg.TemplatesDir, "layout.html")

	templates := make(map[string]*template.Template, len(pageTemplates))
	for name, file := range pageTemplates {
		t, err := template.New(filepath.Base(layoutPath)).Funcs(templateFuncs).
			ParseFiles(layoutPath, filepath.Join(cfg.TemplatesDir, file))
		if err != nil {
			return nil, err
		}
		templates[name] = t
	}

	return &Server{
		cfg:       cfg,
		store:     st,
		engine:    engine,
		templates: templates,
		log:       log,
	}, nil
}

// Handler собирает net/http.ServeMux со всеми маршрутами приложения.
// Используется синтаксис маршрутов Go 1.22+ (метод + {wildcard} прямо в
// шаблоне пути) — отдельный роутер-пакет для проекта такого размера не
// нужен.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Страницы (HTML)
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /tests", s.handleTestsList)
	mux.HandleFunc("GET /tests/{slug}", s.handleQuizPage)
	mux.HandleFunc("GET /prints", s.handlePrintsList)
	mux.HandleFunc("GET /prints/{test}/{result}", s.handlePrintPage)

	// JSON API
	mux.HandleFunc("GET /api/tests/{slug}", s.handleAPIGetTest)
	mux.HandleFunc("POST /api/tests/{slug}/result", s.handleAPISubmitResult)

	// Заглушка для картинки футболки, если у результата нет своей.
	mux.HandleFunc("GET /shirt-image-placeholder", s.handleShirtImagePlaceholder)

	// Технический эндпоинт для проверок живости (docker healthcheck, k8s и т.п.)
	mux.HandleFunc("GET /healthz", s.handleHealthz)

	// Статика: css/js/картинки футболок.
	staticFS := http.FileServer(http.Dir(s.cfg.StaticDir))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFS))

	return withLogging(s.log, mux)
}

// withLogging — простейший middleware логирования запросов. Единственный
// middleware в проекте; если появятся ещё (auth, rate limiting, recover
// от паник) — их стоит собирать в отдельный файл middleware.go и
// применять здесь же цепочкой.
func withLogging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Info("request", "method", r.Method, "path", r.URL.Path)
	})
}
