package store

import "fmt"

// SAdd - добавляет один или несколько элементов в множество
// Если элемент существует - игнорируется
// Если ключ не сущетсвуте - создает новое множество
// Если ключ существует, но не множество - возвращает ошибку
func (s *Store) SAdd(key string, values ...interface{}) error {
	// 1. Проверяем, есть ли что добавить
	if len(values) == 0 {
		return nil
	}

	// 2. Блокируем хранилище для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 3. СЛУЧАЙ 1 - Проверяем, существует ли ключ
	if val, exists := s.data[key]; exists {
		// 4. Проверяем, что тип - множество
		if s.types[key] != TypeSet {
			return fmt.Errorf("key %s exists but is not a set (type: %s)",
				key, s.types[key].String())
		}

		// 5. Получаем множество
		set := val.(*SetValue)

		// 6. Добавляем элементы
		added := 0
		for _, v := range values {
			// Проверяем, есть ли такой элемент
			if _, exists := set.Items[v]; !exists {
				set.Items[v] = struct{}{}
				added++
			}
		}

		// 7. Обновляем статистику
		if added > 0 {
			s.updateStats(SetOp)
		}

		return nil
	}

	// 8. СЛУЧАЙ 2 - Ключа нет - создаем новое множество
	s.data[key] = NewSet()
	s.types[key] = TypeSet
	set := s.data[key].(*SetValue)

	for _, v := range values {
		set.Items[v] = struct{}{}
	}

	s.updateStats(SetOp)
	return nil
}

// SRem - удаляет один или несколько элементов из множества
// Если ключа нет - возвращаем nil
// Если ключ есть, но не множество - ошибка
// Если множество после удаления, становится пустым - удаляем ключ
func (s *Store) SRem(key string, values ...interface{}) error {
	// 1. Проверка на пустой ввод
	if len(values) == 0 {
		return nil
	}

	// 2. Блокируем операцию для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 3. СЛУЧАЙ 1 - Проверка, существует ли ключ
	if val, exists := s.data[key]; exists {
		if s.types[key] != TypeSet {
			return fmt.Errorf("key %s is exists but is not a set (type:%s)",
				key, s.types[key].String())
		}

		// 4. Получаем множество
		set := val.(*SetValue)

		// 5. Считаем, сколько было элементов до удаления
		oldSize := len(set.Items)

		// 6. Удаляем элемент
		for _, v := range values {
			delete(set.Items, v)
		}

		// 7. Считаем, сколько удалили
		removed := oldSize - len(set.Items)

		// 8. Если множество стало пустым - удаляем ключ
		if len(set.Items) == 0 {
			delete(s.data, key)
			delete(s.types, key)
		}

		// 9. Обновляем статистику
		if removed > 0 {
			s.updateStats(DelOp)
		}
	}

	return nil
}

// SIsMember - проверяет, есть ли элемент в множестве
func (s *Store) SIsMember(key string, value interface{}) (bool, error) {
	// 1 Блокируем на чтение
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Проверяем истек ли ключ, если да - удаляем
	if s.isExpiredAndClean(key) {
		return false, nil
	}

	// 3. Проверка, существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return false, nil
	}

	if s.types[key] != TypeSet {
		return false, fmt.Errorf("key %s exists but is not a set (type:%s)",
			key, s.types[key].String())
	}

	// 4. Получаем множество
	set := val.(*SetValue)

	// 5. Проверяем наличие элемента
	_, found := set.Items[value]

	// 6. Обновляем статистику
	s.updateStats(GetOp)

	// 7. Возвращаем результат
	return found, nil
}

// SMembers - возвращает все элементы множества
func (s *Store) SMembers(key string) ([]interface{}, error) {
	// 1. Блокируем для чтения
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Проверка, истек ли ключ, если да - удаляем
	if s.isExpiredAndClean(key) {
		return nil, nil
	}

	// 3. Проверка существования ключа и типа, если ключ существует
	val, exists := s.data[key]
	if !exists {
		return []interface{}{}, nil
	}

	if s.types[key] != TypeSet {
		return nil, fmt.Errorf("key %s exists but not is a set (type:%s)",
			key, s.types[key].String())
	}

	// 3. Получение множества
	set := val.(*SetValue)

	// 4. Создание среза результата
	result := make([]interface{}, 0, len(set.Items))

	// 5. Копирование элементов
	for v := range set.Items {
		result = append(result, v)
	}

	// 6. Обновление статистики
	s.updateStats(GetOp)

	return result, nil
}

// SCard - возвращает количество элементов в множестве
func (s *Store) SCard(key string) (int, error) {
	// 1. Блокируем для чтения
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 2. Проверяем, есть ли ключ (без удаления)
	if s.isExpired(key) {
		return 0, nil
	}

	// 3. Проверяем, существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return 0, nil
	}

	// 3. Проверяем тип
	if s.types[key] != TypeSet {
		return 0, fmt.Errorf("key %s exists but is not a set (type:%s)",
			key, s.types[key].String())
	}

	// 4. Получение множества
	set := val.(*SetValue)

	return len(set.Items), nil
}
