package middleware

import (
	"sync"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/helpers/sql"
)

// FTSPStore — кэш FTSP-запросов по QID (Т7: экземпляр живёт в Middleware,
// а не в пакетном глобале — состояние изолировано на middleware/тест).
type FTSPStore struct {
	lock  sync.RWMutex
	store map[string]sql.FTSP
}

func newFTSPStore() FTSPStore {
	return FTSPStore{store: make(map[string]sql.FTSP)}
}

func (item *FTSPStore) SetFTSP(query *sql.FTSPQuery) {
	id := uuid.NewString()
	query.QID = id

	item.lock.Lock()
	item.store[id] = query.FTSP
	item.lock.Unlock()
}

func (item *FTSPStore) GetFTSP(qid string) (sql.FTSP, bool) {
	item.lock.RLock()
	ftsp, ok := item.store[qid]
	item.lock.RUnlock()
	return ftsp, ok
}

func (item *FTSPStore) GetOrCreateFTSP(query *sql.FTSPQuery) (sql.FTSP, bool) {
	if query.QID == "" {
		item.SetFTSP(query)
		return query.FTSP, true
	}
	return item.GetFTSP(query.QID)
}
