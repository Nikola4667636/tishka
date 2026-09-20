package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/tishka/tishka/internal/models"
)

// JSONStore читает тесты из *.json файлов в указанной директории.
// Все файлы читаются один раз при вызове Load() (обычно — при старте
// сервера) и дальше отдаются из памяти, поэтому чтение теста — это просто
// обращение к map, без дисковых операций на каждый HTTP-запрos.
//
// Чтобы добавить новый тест в продакшене: положить новый файл вида
// data/tests/my-test.json (см. формат в internal/models.Test и в
// data/tests/vibe-check.json как в примере) и перезапустить сервер —
// новый тест появится в списке автоматически, без единой правки кода.
type JSONStore struct {
	dir string

	mu    sync.RWMutex
	tests map[string]models.Test // slug -> тест
	order []string               // порядок slug'ов — чтобы список тестов был стабильным и совпадал с порядком файлов
}

// NewJSONStore создаёт пустой стор, указывающий на директорию dir.
// Данные не читаются до вызова Load().
func NewJSONStore(dir string) *JSONStore {
	return &JSONStore{
		dir:   dir,
		tests: make(map[string]models.Test),
	}
}

// Load читает все *.json файлы из директории, парсит их в models.Test,
// валидирует (см. models.Test.Validate) и складывает в память.
// Если хотя бы один файл невалиден — Load возвращает ошибку и НЕ трогает
// уже загруженные данные, чтобы одна опечатка в новом файле не положила
// уже работающий сервер при вызове Reload().
func (s *JSONStore) Load() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("чтение директории с тестами %q: %w", s.dir, err)
	}

	newTests := make(map[string]models.Test)
	var newOrder []string

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("чтение файла %q: %w", path, err)
		}

		var test models.Test
		if err := json.Unmarshal(data, &test); err != nil {
			return fmt.Errorf("разбор JSON в файле %q: %w", path, err)
		}

		if err := test.Validate(); err != nil {
			return fmt.Errorf("файл %q: %w", path, err)
		}

		if _, exists := newTests[test.Slug]; exists {
			return fmt.Errorf("дублирующийся slug %q (файл %q)", test.Slug, path)
		}

		newTests[test.Slug] = test
		newOrder = append(newOrder, test.Slug)
	}

	if len(newTests) == 0 {
		return fmt.Errorf("в директории %q не найдено ни одного валидного теста (*.json)", s.dir)
	}

	sort.Strings(newOrder) // стабильный порядок независимо от порядка файлов в ОС

	s.mu.Lock()
	s.tests = newTests
	s.order = newOrder
	s.mu.Unlock()

	return nil
}

// ListTests реализует интерфейс Store.
func (s *JSONStore) ListTests(_ context.Context) ([]models.TestSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]models.TestSummary, 0, len(s.order))
	for _, slug := range s.order {
		summaries = append(summaries, s.tests[slug].Summary())
	}
	return summaries, nil
}

// GetTest реализует интерфейс Store.
func (s *JSONStore) GetTest(_ context.Context, slug string) (*models.Test, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	test, ok := s.tests[slug]
	if !ok {
		return nil, ErrNotFound
	}
	return &test, nil
}

// убеждаемся на этапе компиляции, что JSONStore реализует интерфейс Store.
var _ Store = (*JSONStore)(nil)
