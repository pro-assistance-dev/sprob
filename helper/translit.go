package helper

import (
	"bytes"
)

// ////////////////////////////////////////////////////////////////////////////////// //
var baseRuEn = map[string]string{
	"q": "й",
	"w": "ц",
	"e": "у",
	"r": "к",
	"y": "н",
	"t": "е",
	"u": "г",
	"i": "ш",
	"o": "щ",
	"p": "з",
	"[": "х",
	"]": "ъ",
	"a": "ф",
	"s": "ы",
	"d": "в",
	"f": "а",
	"g": "п",
	"h": "р",
	"j": "о",
	"k": "л",
	"l": "д",
	";": "ж",
	`"`: "э",
	"z": "я",
	"x": "ч",
	"c": "с",
	"v": "м",
	"b": "и",
	"n": "т",
	"m": "ь",
	",": "б",
	".": "ю",
	"/": ".",
}

func (h *Helper) TranslitToRu(text string) string {
	if text == "" {
		return ""
	}

	var input = bytes.NewBufferString(text)
	var output = bytes.NewBuffer(nil)

	var rr string
	var ok bool

	for {
		r, _, err := input.ReadRune()
		if err != nil {
			break
		}
		if isRussianChar(r) {
			output.WriteRune(r)
			continue
		}
		rr, ok = baseRuEn[string(r)]
		if ok {
			output.WriteString(rr)
			continue
		}
		rr, ok = baseRuEn[string(r)]
		if ok {
			output.WriteString(rr)
		}
	}
	return output.String()
}

func isRussianChar(r rune) bool {
	switch {
	case r >= 1040 && r <= 1103,
		r == 1105, r == 1025:
		return true
	}
	return false
}
