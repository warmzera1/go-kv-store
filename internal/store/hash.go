package store

import "fmt"

// HSet - устанавливает поле в хеше
// Если ключ не существует - создает новый в хеше
// Если ключ существует, но не хеш - возвращает ошибку
// Если поле уже существовало - перезаписывает значение
func (s *Store) HSet(key string, field string, value interface{}) error {
	// 1. Проверяем, что поле не пустое
	if field == "" {
		return fmt.Errorf("field cannot be empty")
	}

	// 2. Блокируем хранилище для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 3. СЛУЧАЙ 1 - Ключ существует
	if val, exists := s.data[key]; exists {
		if s.types[key] != TypeHash {
			return fmt.Errorf("key %s exists but is not a hash (type: %s)",
				key, s.types[key].String())
		}

		// 4. Приводим к типу
		hash := val.(*HashValue)

		// 5. Устанавливаем значение
		hash.Fields[field] = value

		// 6. Обновляем статистику
		s.updateStats(SetOp)

		return nil
	}

	// 7. СЛУЧАЙ 2 - Ключ не существует
	s.data[key] = NewHash()
	s.types[key] = TypeHash
	hash := s.data[key].(*HashValue)

	// 8. Устанавливаем поле
	hash.Fields[field] = value

	// 9. Обновляем статистику
	s.updateStats(SetOp)

	return nil
}

// HGet - получает значение поля из ключа
// Если ключа нет - возвращает nil, nil (не ошибка)
// Если ключ есть, но не хеш - ошибка
// Если поля нет - возвращаем nil, nil (не ошибка)
func (s *Store) HGet(key string, field string) (interface{}, error) {
	// 1. Проверяем, что поле не пустое
	if field == "" {
		return nil, fmt.Errorf("field cannot be empty")
	}

	// 2. Блокируем для чтения
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 3. Существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return nil, nil
	}

	// 4. Проверяем тип
	if s.types[key] != TypeHash {
		return nil, fmt.Errorf("key %s exists but is not a hash (type: %s)",
			key, s.types[key].String())
	}

	// 5. Получение хеша
	hash := val.(*HashValue)

	// 6. Получение значения поля
	value, exists := hash.Fields[field]
	if !exists {
		return nil, nil
	}

	// 7. Обновляем статистику
	s.updateStats(GetOp)

	return value, nil
}

// HGetAll - возвращает ВСЕ поля и значения хеша
// Если ключа нет - возвращает пустую map, nil
func (s *Store) HGetAll(key string) (map[string]interface{}, error) {
	// 1. Блокируем для чтения
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 2. Существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return map[string]interface{}{}, nil
	}

	if s.types[key] != TypeHash {
		return nil, fmt.Errorf("key %s exists but is not a hash (type: %s)",
			key, s.types[key].String())
	}

	// 3. Получение хеша
	hash := val.(*HashValue)

	// 4. Создание среза результата
	result := make(map[string]interface{}, len(hash.Fields))

	// 5. Копирование элементов
	for k, v := range hash.Fields {
		result[k] = v
	}

	// 6. Обновляем статистику
	s.updateStats(GetOp)

	return result, nil
}

// HDel - удаляет одно или несколько полей из хеша
// Если ключа нет - возвращает nil (ничего не делаем)
// Если ключ есть, но не хеш - ошибка
// Если хеш после удаления становится пустым - удаляем ключ
func (s *Store) HDel(key string, fields ...string) error {
	// 1. Проверка на пустой ввод
	if len(fields) == 0 {
		return nil
	}

	// 2. Блокируем на запись
	s.mu.Lock()
	defer s.mu.Unlock()

	// 3. Проверка, существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return nil
	}

	// 4. Проверка типа
	if s.types[key] != TypeHash {
		return fmt.Errorf("key %s exists but is not a hash (type: %s)",
			key, s.types[key].String())
	}

	// 5. Получаем хеш
	hash := val.(*HashValue)

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
		delete(s.data, key)
		delete(s.types, key)
	}

	// 8. Обновляем статистику
	if removed > 0 {
		s.updateStats(DelOp)
	}

	return nil
}
