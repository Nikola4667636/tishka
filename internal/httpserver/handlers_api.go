package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tishka/tishka/internal/models"
	"github.com/tishka/tishka/internal/quiz"
	"github.com/tishka/tishka/internal/shirtimg"
	"github.com/tishka/tishka/internal/store"
)

// ---------- GET /api/tests/{slug} ----------
//
// Отдаёт "публичную" версию теста в JSON — без тегов и результатов.
// Сегодня фронт её не вызывает (данные уже встроены в HTML страницы
// /tests/{slug}, см. handleQuizPage), но эндпоинт полезен сам по себе:
// например, для будущего мобильного приложения, или чтобы вручную
// посмотреть структуру теста через curl при отладке новых JSON-файлов.

func (s *Server) handleAPIGetTest(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	test, err := s.store.GetTest(r.Context(), slug)
	if err != nil {
		s.respondJSONTestLookupError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, test.Public())
}

// ---------- POST /api/tests/{slug}/result ----------

type submitResultRequest struct {
	Answers map[string]string `json:"answers"`
}

type submitResultResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Quote       string `json:"quote"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

func (s *Server) handleAPISubmitResult(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	test, err := s.store.GetTest(r.Context(), slug)
	if err != nil {
		s.respondJSONTestLookupError(w, err)
		return
	}

	var req submitResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	result, err := s.engine.Compute(test, quiz.Answers(req.Answers))
	if err != nil {
		switch {
		case errors.Is(err, quiz.ErrIncompleteAnswers):
			writeJSONError(w, http.StatusBadRequest, "Отвечены не все вопросы")
		case errors.Is(err, quiz.ErrUnknownOption), errors.Is(err, quiz.ErrUnknownQuestion):
			writeJSONError(w, http.StatusBadRequest, "Некорректный ответ")
		default:
			s.log.Error("compute result", "slug", slug, "error", err)
			writeJSONError(w, http.StatusInternalServerError, "Внутренняя ошибка")
		}
		return
	}

	resp := submitResultResponse{
		ID:          result.ID,
		Name:        result.Name,
		Quote:       result.Quote,
		Description: result.Description,
		ImageURL:    s.shirtImageURL(*result),
	}
	writeJSON(w, http.StatusOK, resp)
}

// shirtImageURL возвращает URL картинки футболки для результата.
// Если у результата задан Image и файл реально существует в StaticDir —
// отдаём ссылку на него. Иначе — ссылку на эндпоинт-заглушку.
func (s *Server) shirtImageURL(result models.Result) string {
	if result.Image != "" {
		clean := filepath.Clean(result.Image)
		// защита от "shirts/../../etc/passwd" и абсолютных путей
		if !strings.HasPrefix(clean, "..") && !filepath.IsAbs(clean) {
			full := filepath.Join(s.cfg.StaticDir, clean)
			if info, err := os.Stat(full); err == nil && !info.IsDir() {
				return "/static/" + filepath.ToSlash(clean)
			}
		}
	}
	return "/shirt-image-placeholder"
}

// ---------- GET /shirt-image-placeholder ----------
//
// Отдаёт SVG-заглушку для случаев, когда у результата нет своей картинки.
// Отдельный эндпоинт (а не файл в static) — чтобы заглушка всегда была
// доступна, даже если static-директория настроена нестандартно.

func (s *Server) handleShirtImagePlaceholder(w http.ResponseWriter, r *http.Request) {
	svg := shirtimg.Placeholder(s.cfg.BrandName)
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(svg)
}

// ---------- GET /healthz ----------

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- json helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) respondJSONTestLookupError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "Внутренняя ошибка"

	if errors.Is(err, store.ErrNotFound) {
		status = http.StatusNotFound
		message = "Тест не найден"
	} else {
		s.log.Error("get test", "error", err)
	}

	writeJSONError(w, status, message)
}
