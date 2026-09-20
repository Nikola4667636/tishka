package httpserver

import "html/template"

// templateFuncs — функции, доступные внутри HTML-шаблонов. Единое место
// для таких хелперов — если понадобится ещё один (форматирование даты,
// склонение другого слова) добавлять сюда же.
var templateFuncs = template.FuncMap{
	"pluralize": pluralizeRu,
}

// pluralizeRu возвращает нужную форму слова для числа n по правилам
// русского языка (1 вопрос, 2 вопроса, 5 вопросов, 11 вопросов, 21 вопрос...).
// one — форма для 1, few — для 2–4, many — для всех остальных случаев.
func pluralizeRu(n int, one, few, many string) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}

	mod10 := abs % 10
	mod100 := abs % 100

	switch {
	case mod10 == 1 && mod100 != 11:
		return one
	case mod10 >= 2 && mod10 <= 4 && !(mod100 >= 12 && mod100 <= 14):
		return few
	default:
		return many
	}
}
