package store

import (
	"sync"
	"time"
)

// Основная структура хранилища
type Store struct {
	data    map[string]interface{} // Основное хранилище ключ - значение
	types   map[string]DataType    // Типы ключей
	expiry  map[string]time.Time   // Ключ -> время истечения
	mu      sync.RWMutex           // Для синхронизации доступа к data
	stats   Stats                  // Статистика операций
	statsMu sync.Mutex             // Отдельный мьютекс для статистики
}

// Создаем новое хранилище
func New() *Store {
	return &Store{
		data:   make(map[string]interface{}),
		types:  make(map[string]DataType),
		expiry: make(map[string]time.Time),
	}
}
