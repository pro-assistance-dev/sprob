package utilHelper

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
	"1": "1",
	"2": "2",
	"3": "3",
	"4": "4",
	"5": "5",
	"6": "6",
	"7": "7",
	"8": "8",
	"9": "9",
	"0": "0",
}

func (h *UtilHelper) TranslitToRu(text string) string {
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


func (h *UtilHelper) TranslitToEng(text string) string {
	if text == "" {
		return ""
	}

	var input = bytes.NewBufferString(text)
	var output = bytes.NewBuffer(nil)

	var rr string
	// var ok bool

	for {
		r, _, err := input.ReadRune()
		if err != nil {
			break
		}
		if !isRussianChar(r) {
			output.WriteRune(r)
			continue
		}

		key, ok := mapkey(baseRuEn, string(r))
		if ok {
			output.WriteString(key)
			continue
		}
		rr, ok = baseRuEn[string(r)]
		if ok {
			output.WriteString(rr)
		}
	}
	return output.String()
}

func mapkey(m map[string]string, value string) (key string, ok bool) {
	for k, v := range m {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}
