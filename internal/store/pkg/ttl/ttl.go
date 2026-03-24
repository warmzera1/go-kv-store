package ttl

import (
	"fmt"
	"math"
	"time"

	"github.com/warmzera1/kv-store/internal/store/core"
)

type TTLStore struct {
	store *core.Store
}

func New(s *core.Store) *TTLStore {
	return &TTLStore{store: s}
}

// Expire - устанавливает время жизни ключа
// Если ключа нет - возвращает false, nil
// Если TTL установлен успешно - возвращает true, false
// seconds - время жизни в секундах (sec>0)
func (s *TTLStore) Expire(key string, seconds int) (bool, error) {
	// 1. Проверяем, что TTL - положительный
	if seconds <= 0 {
		return false, fmt.Errorf("TTL must be positive, got %d", seconds)
	}

	// 2. Блокируем хранилище для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 3. Проверяем, существует ли ключ
	// Если ключа нет - ничего не делаем
	_, exists := s.store.Data[key]
	if !exists {
		return false, nil
	}

	// 4. Вычисляем время истечения
	// time.Now() - текущее время
	// Add(seconds * time.Second)
	expiresAt := time.Now().Add(time.Duration(seconds) * time.Second)

	// 5. Сохраняем время истечения
	s.store.Expiry[key] = expiresAt

	// 6. Возвращаем true - TTL успешно установлен
	return true, nil
}

// TTL - возвращает время жизни ключа в секундах
// Если ключа нет - возвращает -2, nil
// Если TTL не установлен - возвращает -1, nil
// Если ключ истек - возвращает -2, nil
func (s *TTLStore) TTL(key string) (int, error) {
	// 1. Блокируем для чтения
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	// 2. Проверяем, существует ли ключ
	_, exists := s.store.Data[key]
	if !exists {
		return -2, nil
	}

	// 3. Проверяем истек ли ключ (без удаления)
	if s.IsExpired(key) {
		return -2, nil
	}

	// 4. Проверяем, есть ли ключ в expiry
	// Если есть, получаем значение
	expiresAt, exists := s.store.Expiry[key]
	if !exists {
		return -1, nil
	}

	// 5. Вычисляем остаток времени
	remaining := time.Until(expiresAt)

	if remaining <= 0 {
		return -2, nil
	}

	// 6. Получаем секунды
	seconds := remaining.Seconds()

	return int(math.Ceil(seconds)), nil
}

// Persist - удаляет TTL у ключа
// Если ключа нет - возвращает false, nil
// Если TTL не был установлен - возвращает false, nil
// Если TTL был успешно удален - возвращает true, nil
func (s *TTLStore) Persist(key string) (bool, error) {
	// 1. Блокируем для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 2. Проверяем, существует ли ключ
	_, exists := s.store.Data[key]
	if !exists {
		return false, nil
	}

	// 3. Проверяем, есть ли TTL в expiry
	_, exists = s.store.Expiry[key]
	if !exists {
		return false, nil
	}

	// 4. Удаляем ключ из expiry
	delete(s.store.Expiry, key)

	return true, nil
}

// isExpired - проверяет, истек ли ключ (без удаления)
func (s *TTLStore) IsExpired(key string) bool {
	// Получаем время истечения, если установлено
	expiresAt, hasExpiry := s.store.Expiry[key]
	if !hasExpiry {
		// Нет TTL - ключ не может истечь
		return false
	}

	// Сравниваем текущее время с временем истечения
	return time.Now().After(expiresAt)
}

// isExpiredAndClean - проверяет истек ли ключ, если да - удаляет
func (s *TTLStore) IsExpiredAndClean(key string) bool {
	// Получаем время истечения, если установлено
	expiresAt, hasExpiry := s.store.Expiry[key]
	if !hasExpiry {
		return false
	}

	// Проверяем, наступило ли время истечения
	if time.Now().After(expiresAt) {
		// Ключ истек - удаляем его из всех хранилищ
		delete(s.store.Data, key)
		delete(s.store.Types, key)
		delete(s.store.Expiry, key)
		return true
	}

	// Ключ еще не истек
	return false
}

// CleanExpiredKeys - удаляет все истекшие ключи
func (s *TTLStore) CleanExpiredKeys() {
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	now := time.Now()
	for key, expiresAt := range s.store.Expiry {
		if now.After(expiresAt) {
			delete(s.store.Data, key)
			delete(s.store.Types, key)
			delete(s.store.Expiry, key)
		}
	}
}

// StartTTLCleaner - запускает фоновую горутину для удаления истекших ключей
func (s *TTLStore) StartTTLCleaner(interval time.Duration) {

	// 1. Запуска горутину
	// go - запусти эту функцию параллельно, не жди
	go func() {

		// 2. Создаем тикер (будильник)
		ticker := time.NewTicker(interval)

		// 3. Гарантирует остановку тикера
		defer ticker.Stop()

		// 4. Бесконечный цикл
		for range ticker.C {

			// Каждую секунду выполняем
			s.CleanExpiredKeys()
		}
	}()
}
