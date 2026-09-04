// Package analytics — общие агрегаты для дашбордов (А5.3 rdkb/TASKS.md):
// TTL-кэш агрегатов (счётчики/распределения меняются редко, каждый дашборд
// тянет несколько запросов по большим таблицам) + хелперы «серий» для ответов
// и выгрузок. Вынесено из rdkb (map/hr handlers/analytics — код был идентичен).
package analytics

import (
	"sync"
	"time"
)

// Cache — TTL-кэш агрегатов по ключу дашборда.
// Проектный сервис держит один экземпляр: агрегаты кэшируются на TTL,
// «живые» данные (журналы и т.п.) запрашиваются напрямую, минуя кэш.
type Cache struct {
	ttl time.Duration
	mu  sync.Mutex
	m   map[string]cacheEntry
}

type cacheEntry struct {
	at   time.Time
	data interface{}
}

// NewCache создаёт кэш с TTL (например, 60 * time.Second).
func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl, m: map[string]cacheEntry{}}
}

// Get возвращает значение, если оно есть и не протухло.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[key]
	if !ok || time.Since(e.at) > c.ttl {
		return nil, false
	}
	return e.data, true
}

// Set кладёт значение в кэш (сбрасывает TTL).
func (c *Cache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = cacheEntry{at: time.Now(), data: data}
}

// GetOrLoad возвращает закэшированное значение либо вычисляет через load
// и кэширует результат (ошибки не кэшируются).
func (c *Cache) GetOrLoad(key string, load func() (interface{}, error)) (interface{}, error) {
	if v, ok := c.Get(key); ok {
		return v, nil
	}
	data, err := load()
	if err != nil {
		return nil, err
	}
	c.Set(key, data)
	return data, nil
}

// Reset очищает кэш (например, после массовых изменений).
func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = map[string]cacheEntry{}
}
