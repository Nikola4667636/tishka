// Package quiz содержит логику подсчёта результата теста. Она намеренно
// вынесена отдельно от HTTP-слоя (internal/httpserver) и от хранилища
// (internal/store): сюда можно писать unit-тесты без поднятия сервера,
// и здесь же в будущем менять сам алгоритм подбора (см. комментарий
// к Engine ниже), не трогая ни HTTP, ни формат данных.
package quiz

import (
	"errors"
	"fmt"

	"github.com/tishka/tishka/internal/models"
)

// ErrIncompleteAnswers возвращается, если пользователь ответил не на все
// вопросы теста.
var ErrIncompleteAnswers = errors.New("отвечены не все вопросы теста")

// ErrUnknownQuestion / ErrUnknownOption — пользователь прислал id вопроса
// или варианта ответа, которого нет в тесте (испорченный запрос или
// рассинхронизация фронта с данными теста).
var (
	ErrUnknownQuestion = errors.New("неизвестный вопрос")
	ErrUnknownOption   = errors.New("неизвестный вариант ответа")
)

// Answers — ответы пользователя: questionID -> optionID.
type Answers map[string]string

// Engine считает результат теста по ответам пользователя.
//
// Текущий алгоритм — "какой тег встречался чаще всего в выбранных
// вариантах, того и результат" (см. Score). Он специально максимально
// простой для MVP. Если в будущем понадобится более сложная логика
// (например, взвешенные теги, несколько результатов сразу, или ML-модель
// на эмбеддингах ответов) — нужно поменять реализацию Engine.Compute,
// сохранив сигнатуру интерфейса, и весь остальной код (HTTP-хендлеры)
// не заметит разницы.
type Engine struct{}

// NewEngine создаёт движок подсчёта результатов.
func NewEngine() Engine {
	return Engine{}
}

// Compute считает финальный результат теста по ответам пользователя.
func (Engine) Compute(test *models.Test, answers Answers) (*models.Result, error) {
	if len(answers) < len(test.Questions) {
		return nil, ErrIncompleteAnswers
	}

	tagScores := make(map[string]int)

	for _, q := range test.Questions {
		optionID, answered := answers[q.ID]
		if !answered {
			return nil, fmt.Errorf("%w: вопрос %q", ErrIncompleteAnswers, q.ID)
		}

		option, err := findOption(q, optionID)
		if err != nil {
			return nil, err
		}

		for _, tag := range option.Tags {
			tagScores[tag]++
		}
	}

	return pickResult(test, tagScores)
}

func findOption(q models.Question, optionID string) (models.Option, error) {
	for _, o := range q.Options {
		if o.ID == optionID {
			return o, nil
		}
	}
	return models.Option{}, fmt.Errorf("%w: %q в вопросе %q", ErrUnknownOption, optionID, q.ID)
}

// pickResult выбирает результат с максимальным набранным счётом по тегу.
// При равенстве счётов побеждает результат, который раньше объявлен в
// JSON — это делает выбор детерминированным (одинаковые ответы всегда
// дают одинаковый результат), что важно и для тестируемости, и чтобы
// пользователь не удивлялся, почему при повторном прохождении с теми же
// ответами выпала другая футболка.
func pickResult(test *models.Test, tagScores map[string]int) (*models.Result, error) {
	var best *models.Result
	bestScore := -1

	for i := range test.Results {
		r := test.Results[i]
		score := tagScores[r.Tag]
		if score > bestScore {
			bestScore = score
			best = &test.Results[i]
		}
	}

	if best == nil {
		// Такого не должно случиться благодаря models.Test.Validate,
		// но проверяем явно вместо того, чтобы вернуть нулевой Result.
		return nil, fmt.Errorf("не удалось подобрать результат для теста %q", test.Slug)
	}

	return best, nil
}
