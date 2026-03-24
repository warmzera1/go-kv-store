package hash

import (
	"fmt"

	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

type HashStore struct {
	store *core.Store
	ttl   ttl.TTLInterface
}

func New(s *core.Store, t ttl.TTLInterface) *HashStore {
	return &HashStore{
		store: s,
		ttl:   t,
	}
}

// HSet - устанавливает поле в хеше
// Если ключ не существует - создает новый в хеше
// Если ключ существует, но не хеш - возвращает ошибку
// Если поле уже существовало - перезаписывает значение
func (s *HashStore) HSet(key string, field string, value interface{}) error {
	// 1. Проверяем, что поле не пустое
	if field == "" {
		return fmt.Errorf("field cannot be empty")
	}

	// 2. Блокируем хранилище для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 3. СЛУЧАЙ 1 - Ключ существует
	if val, exists := s.store.Data[key]; exists {
		if s.store.Types[key] != core.TypeHash {
			return fmt.Errorf("key %s exists but is not a hash (type: %s)",
				key, s.store.Types[key].String())
		}

		// 4. Приводим к типу
		hash := val.(*core.HashValue)

		// 5. Устанавливаем значение
		hash.Fields[field] = value

		// 6. Обновляем статистику
		s.store.UpdateStats(core.SetOp)

		return nil
	}

	// 7. СЛУЧАЙ 2 - Ключ не существует
	s.store.Data[key] = core.NewHash()
	s.store.Types[key] = core.TypeHash
	hash := s.store.Data[key].(*core.HashValue)

	// 8. Устанавливаем поле
	hash.Fields[field] = value

	// 9. Обновляем статистику
	s.store.UpdateStats(core.SetOp)

	return nil
}

// HGet - получает значение поля из ключа
// Если ключа нет - возвращает nil, nil (не ошибка)
// Если ключ есть, но не хеш - ошибка
// Если поля нет - возвращаем nil, nil (не ошибка)
func (s *HashStore) HGet(key string, field string) (interface{}, error) {
	// 1. Проверяем, что поле не пустое
	if field == "" {
		return nil, fmt.Errorf("field cannot be empty")
	}

	// 2. Блокируем для чтения
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 3. Проверяем истечение и удаляем, если нужно
	if s.ttl.IsExpiredAndClean(key) {
		return nil, nil
	}

	// 4. Существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return nil, nil
	}

	// 5. Проверяем тип
	if s.store.Types[key] != core.TypeHash {
		return nil, fmt.Errorf("key %s exists but is not a hash (type: %s)",
			key, s.store.Types[key].String())
	}

	// 6. Получение хеша
	hash := val.(*core.HashValue)

	// 7. Получение значения поля
	value, exists := hash.Fields[field]
	if !exists {
		return nil, nil
	}

	// 8. Обновляем статистику
	s.store.UpdateStats(core.GetOp)

	return value, nil
}

// HGetAll - возвращает ВСЕ поля и значения хеша
// Если ключа нет - возвращает пустую map, nil
func (s *HashStore) HGetAll(key string) (map[string]interface{}, error) {
	// 1. Блокируем для чтения
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// Проверяем истечение и удаляем, если нужно
	if s.ttl.IsExpiredAndClean(key) {
		return map[string]interface{}{}, nil
	}

	// 2. Существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return map[string]interface{}{}, nil
	}

	if s.store.Types[key] != core.TypeHash {
		return nil, fmt.Errorf("key %s exists but is not a hash (type: %s)",
			key, s.store.Types[key].String())
	}

	// 3. Получение хеша
	hash := val.(*core.HashValue)

	// 4. Создание среза результата
	result := make(map[string]interface{}, len(hash.Fields))

	// 5. Копирование элементов
	for k, v := range hash.Fields {
		result[k] = v
	}

	// 6. Обновляем статистику
	s.store.UpdateStats(core.GetOp)

	return result, nil
}

// HDel - удаляет одно или несколько полей из хеша
// Если ключа нет - возвращает nil (ничего не делаем)
// Если ключ есть, но не хеш - ошибка
// Если хеш после удаления становится пустым - удаляем ключ
func (s *HashStore) HDel(key string, fields ...string) error {
	// 1. Проверка на пустой ввод
	if len(fields) == 0 {
		return nil
	}

	// 2. Блокируем на запись
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 3. Проверка, существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return nil
	}

	// 4. Проверка типа
	if s.store.Types[key] != core.TypeHash {
		return fmt.Errorf("key %s exists but is not a hash (type: %s)",
			key, s.store.Types[key].String())
	}

	// 5. Получаем хеш
	hash := val.(*core.HashValue)

	// 6. Удаление полей (считаем, сколько удалили)
	removed := 0
	for _, field := range fields {
		if _, exists := hash.Fields[field]; exists {
			delete(hash.Fields, field)
			removed++
		}
	}

	// 7. Если хеш стал пустым - удаляем ключ
	if len(hash.Fields) == 0 {
		delete(s.store.Data, key)
		delete(s.store.Types, key)
	}

	// 8. Обновляем статистику
	if removed > 0 {
		s.store.UpdateStats(core.DelOp)
	}

	return nil
}
