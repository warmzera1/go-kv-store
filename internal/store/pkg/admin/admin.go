package admin

import (
	"fmt"
	"sync"

	"github.com/warmzera1/kv-store/internal/store/core"
)

type AdminStore struct {
	store         *core.Store
	bgSaveLock    sync.Mutex
	bgSaveRunning bool
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

// BGSave - фоновое сохранение
func (a *AdminStore) BGSave(filename string) error {
	// 1. Проверяем, не идет ли уже фоновое сохранение
	a.bgSaveLock.Lock()
	if a.bgSaveRunning {
		a.bgSaveLock.Unlock()
		return fmt.Errorf("Background save already in progress")
	}
	a.bgSaveRunning = true
	a.bgSaveLock.Unlock()

	// 2. Запускаем горутину для фонового состония
	go func() {
		defer func() {
			a.bgSaveLock.Lock()
			a.bgSaveRunning = false
			a.bgSaveLock.Unlock()
		}()

		// 2.1 Сохраняем снапшот
		if err := a.store.SaveSnapshot(filename); err != nil {
			fmt.Printf("SAVE error: %v\n", err)
		} else {
			fmt.Printf("SAVE completed: %s\n", filename)
		}
	}()

	return nil
}
