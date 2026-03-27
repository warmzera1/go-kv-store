package admin

import (
	"fmt"
	"sync"
	"time"

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

// Flushdb - очистить kv-store
func (a *AdminStore) FlushDB() error {
	// 1. Блокируем хранилище для записи
	a.store.Mu.Lock()
	defer a.store.Mu.Unlock()

	// 2. Очищаем все данные из хранилища
	a.store.Data = make(map[string]interface{})
	a.store.Expiry = make(map[string]time.Time)
	a.store.Types = make(map[string]core.DataType)

	// 3. Сбрасываем статистику с защитой
	a.store.StatsMu.Lock()
	a.store.Stats = core.Stats{}
	a.store.StatsMu.Unlock()

	return nil
}
