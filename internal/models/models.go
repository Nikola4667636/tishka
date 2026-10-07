// Package models содержит доменные сущности приложения: тест, вопросы,
// варианты ответов и результаты (футболки). Эти структуры — единственное
// место, которое описывает форму теста, поэтому именно сюда нужно смотреть
// в первую очередь, если хочется поменять формат JSON с тестами.
package models

import "fmt"

// Option — один вариант ответа на вопрос.
// Tags — список "тегов вайба", которые получает пользователь, выбрав этот
// вариант. Именно теги, а не сам вариант, участвуют в подсчёте результата —
// это и есть основной механизм подбора футболки.
type Option struct {
	ID   string   `json:"id"`
	Text string   `json:"text"`
	Tags []string `json:"tags"`
}

// Question — один вопрос теста с набором вариантов ответа.
type Question struct {
	ID      string   `json:"id"`
	Text    string   `json:"text"`
	Options []Option `json:"options"`
}

// Result — один из возможных исходов теста: конкретная "футболка",
// которую предлагаем в конце. Tag должен совпадать с одним из тегов,
// встречающихся в Options вопросов — по этому совпадению движок квиза
// и выбирает, какой Result показать.
//
// Одновременно Result — это "принт" в каталоге /prints: коллекция там —
// сам тест, а карточками внутри коллекции служат именно результаты.
type Result struct {
	ID          string `json:"id"`
	Tag         string `json:"tag"`
	Name        string `json:"name"`
	Quote       string `json:"quote"`       // короткая фраза, которая "печатается" на футболке
	Description string `json:"description"` // 1-2 предложения, почему это подходит пользователю
	// Image — путь к файлу-картинке футболки относительно StaticDir,
	// например "shirts/cozy.png". Если пусто или файла нет — покажем заглушку.
	Image string `json:"image,omitempty"`
	// Tags — дополнительные теги для карточки в каталоге /prints.
	// Не участвуют в подборе результата (для этого есть Tag), нужны
	// только для визуальной навигации: "минимализм", "ночь", "ирония".
	Tags []string `json:"tags,omitempty"`
	// Price — цена в рублях. Опционально: если 0, блок с ценой не показываем.
	Price int `json:"price,omitempty"`
}

// Test — тест целиком: вопросы + возможные результаты.
// Для страницы /prints тест выступает как "коллекция": его Slug — slug
// коллекции, Title — название коллекции, Results — карточки принтов.
type Test struct {
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Emoji       string     `json:"emoji"` // маленькая иконка для карточки на странице со списком тестов
	Questions   []Question `json:"questions"`
	Results     []Result   `json:"results"`
}

// TestSummary — облегчённая версия теста для списка на странице /tests,
// без вопросов и результатов (они не нужны, пока пользователь не открыл тест).
type TestSummary struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Emoji       string `json:"emoji"`
	Questions   int    `json:"questions_count"`
}

// Summary превращает полный тест в его облегчённое представление для списка.
func (t Test) Summary() TestSummary {
	return TestSummary{
		Slug:        t.Slug,
		Title:       t.Title,
		Description: t.Description,
		Emoji:       t.Emoji,
		Questions:   len(t.Questions),
	}
}

// PublicOption — вариант ответа в виде, безопасном для отправки в браузер:
// без Tags. Теги — это по сути "ключ" для подсчёта результата, и им не
// место в разметке страницы, которую любой может открыть через
// "Просмотр кода страницы".
type PublicOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// PublicQuestion — вопрос в виде, безопасном для отправки в браузер.
type PublicQuestion struct {
	ID      string         `json:"id"`
	Text    string         `json:"text"`
	Options []PublicOption `json:"options"`
}

// PublicTest — тест в виде, безопасном для отправки в браузер: без Tags
// у вариантов ответа и вообще без Results (иначе можно было бы посмотреть
// в devtools, какая футболка "самая редкая", ещё не пройдя тест).
// Именно эта структура сериализуется в JSON и встраивается в HTML
// страницы /tests/{slug} — см. internal/httpserver.
type PublicTest struct {
	Slug      string           `json:"slug"`
	Title     string           `json:"title"`
	Questions []PublicQuestion `json:"questions"`
}

// Public превращает полный тест в его клиентскую (безопасную) версию.
func (t Test) Public() PublicTest {
	questions := make([]PublicQuestion, 0, len(t.Questions))
	for _, q := range t.Questions {
		options := make([]PublicOption, 0, len(q.Options))
		for _, o := range q.Options {
			options = append(options, PublicOption{ID: o.ID, Text: o.Text})
		}
		questions = append(questions, PublicQuestion{ID: q.ID, Text: q.Text, Options: options})
	}
	return PublicTest{Slug: t.Slug, Title: t.Title, Questions: questions}
}

// Validate проверяет внутреннюю целостность теста: у него должны быть
// вопросы и результаты, у каждого вопроса — минимум 2 варианта, и каждый
// тег, использованный в вариантах ответа, должен быть покрыт хотя бы одним
// результатом (иначе пользователь может "выиграть" тег, для которого нет
// футболки). Вызывается при загрузке данных, чтобы ловить опечатки в JSON
// сразу при старте сервера, а не в рантайме на реальном пользователе.
func (t Test) Validate() error {
	if t.Slug == "" {
		return fmt.Errorf("test: пустой slug")
	}
	if len(t.Questions) == 0 {
		return fmt.Errorf("test %q: нет ни одного вопроса", t.Slug)
	}
	if len(t.Results) == 0 {
		return fmt.Errorf("test %q: нет ни одного результата", t.Slug)
	}

	knownTags := make(map[string]bool)
	for _, r := range t.Results {
		if r.Tag == "" {
			return fmt.Errorf("test %q: результат %q без tag", t.Slug, r.ID)
		}
		knownTags[r.Tag] = true
	}

	usedTags := make(map[string]bool)
	for _, q := range t.Questions {
		if len(q.Options) < 2 {
			return fmt.Errorf("test %q: вопрос %q должен иметь минимум 2 варианта ответа", t.Slug, q.ID)
		}
		for _, o := range q.Options {
			for _, tag := range o.Tags {
				usedTags[tag] = true
			}
		}
	}

	for tag := range usedTags {
		if !knownTags[tag] {
			return fmt.Errorf("test %q: тег %q встречается в ответах, но для него нет результата", t.Slug, tag)
		}
	}

	return nil
}
