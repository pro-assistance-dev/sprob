package middleware

import (
	"sync"

	"github.com/google/uuid"
	"github.com/pro-assistance/pro-assister/helpers/sql"
)

type FTSPStore struct {
	store map[string]sql.FTSP
}

var lock = sync.RWMutex{}

var ftspStore = FTSPStore{store: make(map[string]sql.FTSP)}

func (item FTSPStore) SetFTSP(query *sql.FTSPQuery) {
	id := uuid.NewString()
	// query.FTSP.ID = id
	query.QID = id

	lock.Lock()
	item.store[id] = query.FTSP
	defer lock.Unlock()
}

func (item FTSPStore) GetFTSP(qid string) (sql.FTSP, bool) {
	lock.Lock()
	ftsp, ok := item.store[qid]
	defer lock.Unlock()
	return ftsp, ok
}

func (item FTSPStore) GetOrCreateFTSP(query *sql.FTSPQuery) (sql.FTSP, bool) {
	if query.QID == "" {
		item.SetFTSP(query)
		return query.FTSP, true
	}
	return item.GetFTSP(query.QID)
}
