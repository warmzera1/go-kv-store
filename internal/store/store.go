package store

import (
	"fmt"
	"sync"
)

// Для операций статистики
const (
	SetOp = iota // 0 - операция записи
	GetOp        // 1 - операция чтения
	DelOp        // 2 - операция удаления
)

// Хранит статистику операций
type Stats struct {
	SetCount int
	GetCount int
	DelCount int
}

// Основная структура хранилища
type Store struct {
	data    map[string]interface{} // Основное хранилище ключ - значение
	types   map[string]DataType    // Типы ключей
	mu      sync.RWMutex           // Для синхронизации доступа к data
	stats   Stats                  // Статистика операций
	statsMu sync.Mutex             // Отдельный мьютекс для статистики
}

type StringValue string

// Создаем новое хранилище
func New() *Store {
	return &Store{
		data:  make(map[string]interface{}),
		types: make(map[string]DataType),
	}
}

// updateStats - обновляет статистику операций
func (s *Store) updateStats(op int) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()

	switch op {
	case SetOp:
		s.stats.SetCount++
	case GetOp:
		s.stats.GetCount++
	case DelOp:
		s.stats.DelCount++
	}
}

func (s *Store) Set(key, value string) {
	// 1. Блокируем запись в data
	s.mu.Lock()

	// 2. Откладываем разблокировку
	defer s.mu.Unlock()

	// 3. Сохраняем значение
	s.data[key] = StringValue(value)
	s.data[key] = TypeString

	// 4. Обновляем статистику
	// Используем отдельный мьютекс, чтобы не блокировать данные
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.stats.SetCount++
}

func (s *Store) Get(key string) (string, bool) {
	// 1. Блокируем данные для чтения
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 2. Прочитываем значение из существующей map
	val, exists := s.data[key]
	if !exists {
		return "", false
	}

	// 3. Проверяем, что действительно строка
	if s.types[key] != TypeString {
		return "", false
	}

	strVal, ok := val.(StringValue)
	if !ok {
		return "", false
	}

	// 4. Обновляем статистику
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.stats.GetCount++

	// 5. Возвращаем результат
	return string(strVal), true
}

func (s *Store) Delete(key string) {
	// 1. Блокируем данные для чтения
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Удаляем
	delete(s.data, key)

	// 3. Обновляем статистику
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.stats.DelCount++
}

func (s *Store) Stats() Stats {
	// 1. Блокируем данные для чтения
	s.statsMu.Lock()
	defer s.statsMu.Unlock()

	// 2. Возвращаем копию, чтобы нельзя было изменить оригинал
	return Stats{
		SetCount: s.stats.SetCount,
		GetCount: s.stats.GetCount,
		DelCount: s.stats.DelCount,
	}
}

// LPush - добавляет элемент в НАЧАЛО списка (Left Push)
func (s *Store) LPush(key string, value interface{}) error {
	// 1. Блокируем для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Пытаемся получить существующий список
	if existing, exists := s.data[key]; exists {
		if s.types[key] != TypeList {
			return fmt.Errorf("key %s exists but is not a list (type:%s)", key, s.types[key])
		}

		// 3. Приводим к типу
		list := existing.(*ListValue)

		// 4. Добавляем в начало
		list.Items = append([]interface{}{value}, list.Items...)

		s.updateStats(SetOp)
		return nil
	}

	// 5. Создаем новый список
	s.data[key] = NewList()
	s.types[key] = TypeList
	list := s.data[key].(*ListValue)
	list.Items = []interface{}{value}

	// 6. Обновляем статистику
	s.updateStats(SetOp)
	return nil
}

// RPush - добавляем элемент в конец списка
func (s *Store) RPush(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, exists := s.data[key]; exists {
		if s.types[key] != TypeList {
			return fmt.Errorf("key %s exists but is not a list (type:%s)", key, s.types[key])
		}

		list := existing.(*ListValue)
		list.Items = append(list.Items, value)
		s.updateStats(SetOp)
		return nil
	}

	s.data[key] = NewList()
	s.types[key] = TypeList
	list := s.data[key].(*ListValue)
	list.Items = []interface{}{value}
	s.updateStats(SetOp)
	return nil
}
