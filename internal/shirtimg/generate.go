// Package shirtimg отдаёт картинку-заглушку для случая, когда у результата
// теста ещё нет реального файла футболки. Раньше здесь генерировался SVG
// "на лету" по цветам и цитате из JSON; теперь футболки — это готовые
// файлы (см. web/static/shirts/), а этот пакет нужен только чтобы
// у сайта была осмысленная картинка-заглушка, если файла нет.
//
// Когда появятся реальные принты — достаточно положить файлы в
// web/static/shirts/ и указать путь в поле "image" результата в JSON.
// Заглушка останется на случай, если для какого-то результата картинку
// ещё не нарисовали — сайт при этом не упадёт.
package shirtimg

import (
	"html"
	"strings"
)

// Placeholder возвращает байты SVG-заглушки: нейтральный силуэт футболки
// с подписью "скоро здесь будет твоя футболка". brandLabel — опциональная
// подпись бренда внизу картинки (если пусто — не выводится).
func Placeholder(brandLabel string) []byte {
	var b strings.Builder

	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 360" width="300" height="360" role="img" aria-label="Футболка скоро появится">`)
	b.WriteString(`<rect x="0" y="0" width="300" height="360" rx="24" fill="#F6D9DE"/>`)
	b.WriteString(`<path d="M150,46 C160,46 168,51 173,58 L222,54 L266,96 L222,124 L222,278 L78,278 L78,124 L34,96 L78,54 L127,58 C132,51 140,46 150,46 Z" fill="#C9B79C" stroke="#5C1A1A" stroke-width="3" stroke-linejoin="round"/>`)
	b.WriteString(`<line x1="150" y1="70" x2="150" y2="270" stroke="#5C1A1A" stroke-width="1" stroke-opacity="0.15"/>`)
	b.WriteString(`<rect x="98" y="140" width="104" height="70" rx="10" fill="#FFFFFF" fill-opacity="0.92"/>`)
	b.WriteString(`<text x="150" y="170" text-anchor="middle" font-family="Georgia, serif" font-style="italic" font-size="14" fill="#5C1A1A">скоро здесь будет</text>`)
	b.WriteString(`<text x="150" y="192" text-anchor="middle" font-family="Georgia, serif" font-style="italic" font-size="14" fill="#5C1A1A">твоя футболка</text>`)

	if brandLabel != "" {
		b.WriteString(`<text x="150" y="335" text-anchor="middle" font-family="Georgia, serif" font-size="16" letter-spacing="3" fill="#5C1A1A">`)
		b.WriteString(html.EscapeString(strings.ToUpper(brandLabel)))
		b.WriteString(`</text>`)
	}

	b.WriteString(`</svg>`)

	return []byte(b.String())
}
