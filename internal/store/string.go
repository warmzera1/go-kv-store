package store

// func (s *Store) Set(key, value string) {
// 	// 1. Блокируем запись в data
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	// 3. Сохраняем значение
// 	s.data[key] = StringValue(value)
// 	s.types[key] = TypeString

// 	// 4. Обновляем статистику
// 	s.updateStats(SetOp)
// }

// func (s *Store) Get(key string) (string, bool) {
// 	// 1. Блокируем данные для чтения
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	// 2. Проверяем, истек ли ключ, если да - удаляем
// 	if s.isExpiredAndClean(key) {
// 		return "", false
// 	}

// 	// 3. Проверяем, существует ли ключ и проверяем тип
// 	val, exists := s.data[key]
// 	if !exists || s.types[key] != TypeString {
// 		return "", false
// 	}

// 	// 4. Обновляем статистику
// 	s.updateStats(GetOp)

// 	// 5. Возвращаем результат
// 	return string(val.(StringValue)), true
// }

// func (s *Store) Delete(key string) {
// 	// 1. Блокируем данные для чтения
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	// 2. Удаляем
// 	delete(s.data, key)
// 	delete(s.types, key)

// 	// 3. Обновляем статистику
// 	s.updateStats(DelOp)
// }
