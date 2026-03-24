package core

import (
	"sync"
	"time"
)

// Основная структура хранилища
type Store struct {
	Data    map[string]interface{} // Основное хранилище ключ - значение
	Types   map[string]DataType    // Типы ключей
	Expiry  map[string]time.Time   // Ключ -> время истечения
	Mu      sync.RWMutex           // Для синхронизации доступа к data
	Stats   Stats                  // Статистика операций
	StatsMu sync.Mutex             // Отдельный мьютекс для статистики
}

// Создаем новое хранилище
func New() *Store {
	return &Store{
		Data:   make(map[string]interface{}),
		Types:  make(map[string]DataType),
		Expiry: make(map[string]time.Time),
	}
}
