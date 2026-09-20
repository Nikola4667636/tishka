package httpserver

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	"github.com/tishka/tishka/internal/models"
	"github.com/tishka/tishka/internal/store"
)

// baseData — поля, нужные всем страницам (шапка, футер, <title>).
// Каждая страница встраивает baseData в свою собственную структуру
// данных и добавляет то, что нужно только ей.
type baseData struct {
	Brand     string
	PageTitle string
}

func (s *Server) newBaseData(pageTitle string) baseData {
	return baseData{Brand: s.cfg.BrandName, PageTitle: pageTitle}
}

// ---------- GET / ----------

type indexData struct {
	baseData
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := indexData{baseData: s.newBaseData("Кастомные футболки по вайбу")}
	s.render(w, "index", data)
}

// ---------- GET /tests ----------

type testsData struct {
	baseData
	Tests []models.TestSummary
}

func (s *Server) handleTestsList(w http.ResponseWriter, r *http.Request) {
	tests, err := s.store.ListTests(r.Context())
	if err != nil {
		s.log.Error("list tests", "error", err)
		http.Error(w, "Не получилось загрузить список тестов", http.StatusInternalServerError)
		return
	}

	data := testsData{
		baseData: s.newBaseData("Тесты"),
		Tests:    tests,
	}
	s.render(w, "tests", data)
}

// ---------- GET /tests/{slug} ----------

type quizData struct {
	baseData
	Test     models.PublicTest
	TestJSON template.JS
}

func (s *Server) handleQuizPage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	test, err := s.store.GetTest(r.Context(), slug)
	if err != nil {
		s.respondTestLookupError(w, err)
		return
	}

	publicTest := test.Public()

	jsonBytes, err := json.Marshal(publicTest)
	if err != nil {
		s.log.Error("marshal public test", "slug", slug, "error", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	data := quizData{
		baseData: s.newBaseData(test.Title),
		Test:     publicTest,
		// jsonBytes собран из данных, которые мы сами же валидировали при
		// загрузке (models.Test.Validate), поэтому безопасно пометить их
		// как доверенный JS и вставить как есть внутрь <script>.
		TestJSON: template.JS(jsonBytes),
	}
	s.render(w, "quiz", data)
}

// respondTestLookupError отвечает 404, если тест не найден, и 500 — на
// любую другую ошибку стора. Общая для всех хендлеров, которым нужен
// конкретный тест по slug.
func (s *Server) respondTestLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "Тест не найден", http.StatusNotFound)
		return
	}
	s.log.Error("get test", "error", err)
	http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
}

// render выполняет именованный шаблон страницы. Общая точка рендера для
// всех HTML-хендлеров — если понадобится добавить общий пост-процессинг
// (например, security-заголовки для HTML-ответов), делать это тут.
func (s *Server) render(w http.ResponseWriter, name string, data any) {
	t, ok := s.templates[name]
	if !ok {
		http.Error(w, "Шаблон не найден: "+name, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		s.log.Error("execute template", "name", name, "error", err)
	}
}
