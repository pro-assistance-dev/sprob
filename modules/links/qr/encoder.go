package qr

// QR-энкодер (ISO/IEC 18004): byte mode, ECC M, версии 1–10 (до 213 байт).
// Порт клиентского генератора rdkb/map (client/src/qr/encoder.ts) — та же
// реализация, чтобы напечатанные старые и новые наклейки были одинаковыми.

import (
	"errors"
	"unicode/utf8"
)

const eccLevelM = 0

type versionSpec struct {
	version    int
	ecPerBlock int
	groups     [][2]int
	align      []int
}

// Таблица 9 ISO/IEC 18004, уровень M.
var versionsM = []versionSpec{
	{1, 10, [][2]int{{1, 16}}, nil},
	{2, 16, [][2]int{{1, 28}}, []int{6, 18}},
	{3, 26, [][2]int{{1, 44}}, []int{6, 22}},
	{4, 18, [][2]int{{2, 32}}, []int{6, 26}},
	{5, 24, [][2]int{{2, 43}}, []int{6, 30}},
	{6, 16, [][2]int{{4, 27}}, []int{6, 34}},
	{7, 18, [][2]int{{4, 31}}, []int{6, 22, 38}},
	{8, 22, [][2]int{{2, 38}, {2, 39}}, []int{6, 24, 42}},
	{9, 22, [][2]int{{3, 36}, {2, 37}}, []int{6, 26, 46}},
	{10, 26, [][2]int{{4, 43}, {1, 44}}, []int{6, 28, 50}},
}

var errTooLong = errors.New("qr: текст слишком длинный (лимит ~213 байт для версии 10-M)")

func dataCapacityBits(spec versionSpec) int {
	total := 0
	for _, g := range spec.groups {
		total += g[0] * g[1]
	}
	return total * 8
}

// ── GF(256) для Рида—Соломона ────────────────────────────────────────────────

var (
	gfExp [512]byte
	gfLog [256]byte
)

func init() {
	x := 1
	for i := 0; i < 255; i++ {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x11d
		}
	}
	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[int(gfLog[a])+int(gfLog[b])]
}

// rsGenerator — коэффициенты генератора для n контрольных байт.
func rsGenerator(n int) []byte {
	poly := []byte{1}
	for i := 0; i < n; i++ {
		next := make([]byte, len(poly)+1)
		for j := range poly {
			next[j] ^= poly[j]
			next[j+1] ^= gfMul(poly[j], gfExp[i])
		}
		poly = next
	}
	return poly[1:]
}

func rsEncode(data []byte, ecLen int) []byte {
	gen := rsGenerator(ecLen)
	res := make([]byte, ecLen)
	for _, b := range data {
		factor := b ^ res[0]
		copy(res, res[1:])
		res[len(res)-1] = 0
		for i := range gen {
			res[i] ^= gfMul(gen[i], factor)
		}
	}
	return res
}

// ── BitBuffer ───────────────────────────────────────────────────────────────

type bitBuffer struct{ bits []int }

func (b *bitBuffer) put(value, length int) {
	for i := length - 1; i >= 0; i-- {
		b.bits = append(b.bits, (value>>i)&1)
	}
}

func (b *bitBuffer) len() int {
	return len(b.bits)
}

func pickVersion(byteLen int) (versionSpec, error) {
	for _, spec := range versionsM {
		lenBits := 8
		if spec.version >= 10 {
			lenBits = 16
		}
		if 4+lenBits+byteLen*8 <= dataCapacityBits(spec) {
			return spec, nil
		}
	}
	return versionSpec{}, errTooLong
}

func buildDataCodewords(text string, spec versionSpec) []byte {
	payload := []byte(text)
	buf := &bitBuffer{}
	buf.put(0b0100, 4)
	if spec.version >= 10 {
		buf.put(len(payload), 16)
	} else {
		buf.put(len(payload), 8)
	}
	for _, b := range payload {
		buf.put(int(b), 8)
	}

	capacity := dataCapacityBits(spec)
	term := 4
	if capacity-buf.len() < term {
		term = capacity - buf.len()
	}
	buf.put(0, term)
	for buf.len()%8 != 0 {
		buf.put(0, 1)
	}

	codewords := make([]byte, 0, capacity/8)
	for i := 0; i < buf.len(); i += 8 {
		var v byte
		for j := 0; j < 8; j++ {
			v = v<<1 | byte(buf.bits[i+j])
		}
		codewords = append(codewords, v)
	}
	pad := 0
	for len(codewords)*8 < capacity {
		if pad%2 == 0 {
			codewords = append(codewords, 0xec)
		} else {
			codewords = append(codewords, 0x11)
		}
		pad++
	}
	return codewords
}

func interleaveWithEcc(spec versionSpec, data []byte) []byte {
	type block struct {
		data []byte
		ec   []byte
	}
	blocks := make([]block, 0, 4)
	offset := 0
	for _, g := range spec.groups {
		for i := 0; i < g[0]; i++ {
			chunk := data[offset : offset+g[1]]
			offset += g[1]
			blocks = append(blocks, block{data: chunk, ec: rsEncode(chunk, spec.ecPerBlock)})
		}
	}
	maxData := 0
	for _, b := range blocks {
		if len(b.data) > maxData {
			maxData = len(b.data)
		}
	}
	result := make([]byte, 0, len(data)+spec.ecPerBlock*len(blocks))
	for i := 0; i < maxData; i++ {
		for _, b := range blocks {
			if i < len(b.data) {
				result = append(result, b.data[i])
			}
		}
	}
	for i := 0; i < spec.ecPerBlock; i++ {
		for _, b := range blocks {
			result = append(result, b.ec[i])
		}
	}
	return result
}

// ── Матрица ─────────────────────────────────────────────────────────────────

func newMatrix(size int) [][]bool {
	m := make([][]bool, size)
	for i := range m {
		m[i] = make([]bool, size)
	}
	return m
}

func functionModules(spec versionSpec) ([][]bool, [][]bool) {
	size := spec.version*4 + 17
	mods := newMatrix(size)
	reserved := newMatrix(size)
	set := func(r, c int, dark bool) {
		mods[r][c] = dark
		reserved[r][c] = true
	}
	reserve := func(r, c int) { reserved[r][c] = true }

	finder := func(row, col int) {
		for r := -1; r <= 7; r++ {
			for c := -1; c <= 7; c++ {
				rr, cc := row+r, col+c
				if rr < 0 || rr >= size || cc < 0 || cc >= size {
					continue
				}
				in := r >= 0 && r <= 6 && c >= 0 && c <= 6
				border := in && (r == 0 || r == 6 || c == 0 || c == 6)
				center := in && r >= 2 && r <= 4 && c >= 2 && c <= 4
				set(rr, cc, border || center)
			}
		}
	}
	finder(0, 0)
	finder(0, size-7)
	finder(size-7, 0)

	a := spec.align
	for _, row := range a {
		for _, col := range a {
			if (row <= 8 && col <= 8) || (row <= 8 && col >= size-9) || (row >= size-9 && col <= 8) {
				continue
			}
			for r := -2; r <= 2; r++ {
				for c := -2; c <= 2; c++ {
					dark := abs(r) == 2 || abs(c) == 2 || (r == 0 && c == 0)
					set(row+r, col+c, dark)
				}
			}
		}
	}

	for i := 8; i <= size-9; i++ {
		dark := i%2 == 0
		set(6, i, dark)
		set(i, 6, dark)
	}

	for i := 0; i <= 8; i++ {
		reserve(8, i)
		reserve(i, 8)
	}
	for i := 0; i < 8; i++ {
		reserve(8, size-1-i)
		reserve(size-1-i, 8)
	}
	if spec.version >= 7 {
		for i := 0; i < 6; i++ {
			for j := 0; j < 3; j++ {
				reserve(size-11+j, i)
				reserve(i, size-11+j)
			}
		}
	}
	return mods, reserved
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func placeData(mods [][]bool, codewords []byte, reserved [][]bool, spec versionSpec) {
	size := spec.version*4 + 17
	bits := make([]int, 0, len(codewords)*8)
	for _, cw := range codewords {
		for i := 7; i >= 0; i-- {
			bits = append(bits, int(cw>>i)&1)
		}
	}
	bitIndex := 0
	upward := true
	for col := size - 1; col > 0; col -= 2 {
		if col == 6 {
			col = 5
		}
		for i := 0; i < size; i++ {
			row := i
			if upward {
				row = size - 1 - i
			}
			for _, c := range []int{col, col - 1} {
				if reserved[row][c] {
					continue
				}
				bit := 0
				if bitIndex < len(bits) {
					bit = bits[bitIndex]
				}
				mods[row][c] = bit == 1
				bitIndex++
			}
		}
		upward = !upward
	}
}

var maskFuncs = []func(r, c int) bool{
	func(r, c int) bool { return (r+c)%2 == 0 },
	func(r, _ int) bool { return r%2 == 0 },
	func(_, c int) bool { return c%3 == 0 },
	func(r, c int) bool { return (r+c)%3 == 0 },
	func(r, c int) bool { return (r/2+c/3)%2 == 0 },
	func(r, c int) bool { return (r*c)%2+(r*c)%3 == 0 },
	func(r, c int) bool { return ((r*c)%2+(r*c)%3)%2 == 0 },
	func(r, c int) bool { return ((r+c)%2+(r*c)%3)%2 == 0 },
}

func applyMask(mods, reserved [][]bool, maskIdx int) [][]bool {
	out := newMatrix(len(mods))
	fn := maskFuncs[maskIdx]
	for r := range mods {
		for c := range mods {
			out[r][c] = mods[r][c]
			if !reserved[r][c] && fn(r, c) {
				out[r][c] = !out[r][c]
			}
		}
	}
	return out
}

// maskPenalty — упрощённые правила 1 и 2 ISO (маска с минимальным штрафом).
func maskPenalty(m [][]bool, size int) int {
	penalty := 0
	runPenalty := func(run int) int {
		if run >= 5 {
			return 3 + (run - 5)
		}
		return 0
	}
	for r := 0; r < size; r++ {
		run := 1
		for c := 1; c < size; c++ {
			if m[r][c] == m[r][c-1] {
				run++
			} else {
				penalty += runPenalty(run)
				run = 1
			}
		}
		penalty += runPenalty(run)
	}
	for c := 0; c < size; c++ {
		run := 1
		for r := 1; r < size; r++ {
			if m[r][c] == m[r-1][c] {
				run++
			} else {
				penalty += runPenalty(run)
				run = 1
			}
		}
		penalty += runPenalty(run)
	}
	for r := 0; r < size-1; r++ {
		for c := 0; c < size-1; c++ {
			// Блок 2×2 одного цвета (правило 2): сравнение каждой клетки с базовой.
			v := m[r][c]
			same := m[r][c+1] == v && m[r+1][c] == v && m[r+1][c+1] == v
			if same {
				penalty += 3
			}
		}
	}
	return penalty
}

var formatECCBits = map[int]int{0: 0b00, 1: 0b01, 2: 0b10, 3: 0b11}

// formatBits — 15-битная форматная информация BCH(15,5) + маска 0x5412.
func formatBits(eccIndex, maskIdx int) int {
	data := formatECCBits[eccIndex]<<3 | maskIdx
	value := data << 10
	for i := 14; i >= 10; i-- {
		if (value>>i)&1 == 1 {
			value ^= 0x537 << (i - 10)
		}
	}
	return (data<<10 | value) ^ 0x5412
}

func versionBits(version int) int {
	value := version << 12
	for i := 17; i >= 12; i-- {
		if (value>>i)&1 == 1 {
			value ^= 0x1f25 << (i - 12)
		}
	}
	return version<<12 | value
}

func placeVersion(mods [][]bool, size, version int) {
	if version < 7 {
		return
	}
	bits := versionBits(version)
	bit := func(i int) bool { return (bits>>i)&1 == 1 }
	for i := 0; i < 18; i++ {
		dark := bit(i)
		row := i / 3
		col := size - 11 + i%3
		mods[row][col] = dark
		mods[col][row] = dark
	}
}

func placeFormat(mods [][]bool, size, eccIndex, maskIdx int) {
	bits := formatBits(eccIndex, maskIdx)
	bit := func(i int) bool { return (bits>>i)&1 == 1 }

	for i := 0; i <= 5; i++ {
		mods[8][i] = bit(14 - i)
	}
	mods[8][7] = bit(14 - 6)
	mods[8][8] = bit(14 - 7)
	mods[7][8] = bit(14 - 8)
	for i := 9; i <= 14; i++ {
		mods[14-i][8] = bit(14 - i)
	}
	for i := 0; i <= 6; i++ {
		mods[size-7+i][8] = bit(8 + i)
	}
	for i := 0; i <= 7; i++ {
		mods[8][size-8+i] = bit(7 - i)
	}
	mods[size-8][8] = true
}

func buildWithMask(text string, maskIdx int) ([][]bool, error) {
	spec, err := pickVersion(len([]byte(text)))
	if err != nil {
		return nil, err
	}
	data := buildDataCodewords(text, spec)
	codewords := interleaveWithEcc(spec, data)
	size := spec.version*4 + 17

	mods, reserved := functionModules(spec)
	placeData(mods, codewords, reserved, spec)
	masked := applyMask(mods, reserved, maskIdx)
	placeVersion(masked, size, spec.version)
	placeFormat(masked, size, eccLevelM, maskIdx)
	return masked, nil
}

// Encode — матрица QR для текста (лучшая маска из 8).
func Encode(text string) (Matrix, error) {
	if !utf8.ValidString(text) {
		return nil, errors.New("qr: текст не в UTF-8")
	}
	var best Matrix
	bestPenalty := 1 << 30
	for m := 0; m < 8; m++ {
		candidate, err := buildWithMask(text, m)
		if err != nil {
			return nil, err
		}
		if p := maskPenalty(candidate, len(candidate)); p < bestPenalty {
			bestPenalty = p
			best = candidate
		}
	}
	return best, nil
}
