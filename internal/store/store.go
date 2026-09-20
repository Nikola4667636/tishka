// Package store отвечает за то, откуда приложение берёт данные тестов.
//
// Сейчас единственная реализация — JSONStore, которая читает *.json файлы
// из папки при старте сервера. Это осознанный выбор для этапа "проверяем
// гипотезу": редактировать JSON и перезапускать сервер намного быстрее,
// чем поднимать базу данных. Когда тестов станет много и их нужно будет
// редактировать не через git, а через админку — нужно будет добавить
// новую реализацию интерфейса Store (например, PostgresStore), и весь
// остальной код (handlers, quiz engine) менять не придётся, потому что
// они зависят только от интерфейса, а не от JSONStore напрямую.
package store

import (
	"context"

	"github.com/tishka/tishka/internal/models"
)

// Store — контракт, которому должен соответствовать любой источник тестов.
type Store interface {
	// ListTests возвращает краткую информацию по всем доступным тестам —
	// используется на странице /tests.
	ListTests(ctx context.Context) ([]models.TestSummary, error)

	// GetTest возвращает тест целиком по его slug (например, "vibe-check").
	// Если теста с таким slug нет — возвращает ErrNotFound.
	GetTest(ctx context.Context, slug string) (*models.Test, error)
}

// ErrNotFound возвращается реализациями Store, когда тест с таким slug
// не найден. Обработчики HTTP проверяют именно эту ошибку через errors.Is,
// чтобы вернуть пользователю 404, а не 500.
var ErrNotFound = notFoundError{}

type notFoundError struct{}

func (notFoundError) Error() string { return "тест не найден" }
