package admin

import "github.com/warmzera1/kv-store/internal/store/core"

type AdminStore struct {
	store *core.Store
}

func New(s *core.Store) *AdminStore {
	return &AdminStore{
		store: s,
	}
}

func (a *AdminStore) SaveSnapshot(filename string) error {
	return a.store.SaveSnapshot(filename)
}

func (a *AdminStore) LoadSnapshot(filename string) error {
	return a.store.LoadSnapshot(filename)
}
