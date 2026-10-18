package qr

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestEncode_DecodesWithRealDetector — барьер, который ловит главную ошибку
// самописного энкодера: матрица «выглядит как QR», но не читается сканером.
// Декодер берётся из окружения (OpenCV в venv): QR_DECODER=/path/to/python.
// Без него тест пропускается — CI без OpenCV не падает, локально проверяем всегда.
func TestEncode_DecodesWithRealDetector(t *testing.T) {
	decoder := os.Getenv("QR_DECODER")
	if decoder == "" {
		t.Skip("QR_DECODER не задан (нужен python с opencv-python-headless)")
	}

	cases := []string{
		"https://links.rdkb-mos.ru/r/abc234",
		"https://links.rdkb-mos.ru/r/zz99mm",
		// Длинная цель: ссылка на карточку помещения с параметрами (213-байтовый лимит).
		"https://map.rdkb-mos.ru/buildings?roomId=4f6a1c2e-9d3b-4c1a-8a77-0e5f6b1d2c33&tab=passport&hl=1",
	}
	for _, text := range cases {
		m, err := Encode(text)
		if err != nil {
			t.Fatalf("Encode(%q): %v", text, err)
		}
		pngPath := filepath.Join(t.TempDir(), "qr.png")
		if err := os.WriteFile(pngPath, m.PNG(512), 0o644); err != nil {
			t.Fatalf("write png: %v", err)
		}

		script := "import cv2,sys;print(cv2.QRCodeDetector().detectAndDecode(cv2.imread(sys.argv[1]))[0])"
		out, err := exec.Command(decoder, "-c", script, pngPath).Output()
		if err != nil {
			t.Fatalf("декодер не запустился (%s): %v", decoder, err)
		}
		got := strings.TrimSpace(string(out))
		if got != text {
			t.Fatalf("декодировано %q, ожидалось %q", got, text)
		}
	}
}
