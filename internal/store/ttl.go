package store

import (
	"fmt"
	"math"
	"time"
)

// Expire - устанавливает время жизни ключа
// Если ключа нет - возвращает false, nil
// Если TTL установлен успешно - возвращает true, false
// seconds - время жизни в секундах (sec>0)
func (s *Store) Expire(key string, seconds int) (bool, error) {
	// 1. Проверяем, что TTL - положительный
	if seconds <= 0 {
		return false, fmt.Errorf("TTL must be positive, got %d", seconds)
	}

	// 2. Блокируем хранилище для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 3. Проверяем, существует ли ключ
	// Если ключа нет - ничего не делаем
	_, exists := s.data[key]
	if !exists {
		return false, nil
	}

	// 4. Вычисляем время истечения
	// time.Now() - текущее время
	// Add(seconds * time.Second)
	expiresAt := time.Now().Add(time.Duration(seconds) * time.Second)

	// 5. Сохраняем время истечения
	s.expiry[key] = expiresAt

	// 6. Возвращаем true - TTL успешно установлен
	return true, nil
}

// TTL - возвращает время жизни ключа в секундах
// Если ключа нет - возвращает -2, nil
// Если TTL не установлен - возвращает -1, nil
// Если ключ истек - возвращает -2, nil
func (s *Store) TTL(key string) (int, error) {
	// 1. Блокируем для чтения
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 2. Проверяем, существует ли ключ
	_, exists := s.data[key]
	if !exists {
		return -2, nil
	}

	// 3. Проверяем, есть ли ключ в expiry
	// Если есть, получаем значение
	expiresAt, exists := s.expiry[key]
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
func (s *Store) Persist(key string) (bool, error) {
	// 1. Блокируем для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Проверяем, существует ли ключ
	_, exists := s.data[key]
	if !exists {
		return false, nil
	}

	// 3. Проверяем, есть ли TTL в expiry
	_, exists = s.expiry[key]
	if !exists {
		return false, nil
	}

	// 4. Удаляем ключ из expiry
	delete(s.expiry, key)

	return true, nil
}
