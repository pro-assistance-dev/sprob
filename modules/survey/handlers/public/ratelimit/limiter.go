package ratelimit

import (
	"sync"
	"time"
)

// Limiter — простой in-memory лимитер по ключу (IP+маршрут): не более `limit`
// событий за окно `window`. Нужен для защиты публичного приёма ответов от
// накрутки/флуда. In-memory достаточно: сервис одноинстансный, а жёсткая
// точность для антифрода не требуется (дублирование покрывает RespondentHash).
type Limiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

// New — лимитер: `limit` событий за `window`.
func New(limit int, window time.Duration) *Limiter {
	return &Limiter{hits: make(map[string][]time.Time), limit: limit, window: window}
}

// Allow — можно ли выполнить событие с данным ключом. Чистит устаревшие отметки.
func (l *Limiter) Allow(key string) bool {
	if l.limit <= 0 {
		return true
	}
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.limit {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}
