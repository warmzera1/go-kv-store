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

// LPush - добавляет элемент в НАЧАЛО списка
// Если ключ не существует, создает новый список
// Если ключ существует, но не список - возвращает ошибку
func (s *Store) LPush(key string, values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	// 1. Блокируем для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Случай 1 - Ключ уже существует
	if val, exists := s.data[key]; exists {
		if s.types[key] != TypeList {
			return fmt.Errorf("key %s exists but is not a list (type:%s)",
				key, s.types[key].String())
		}

		// 3. Приводим к типу
		list := val.(*ListValue)

		// 4. Добавляем в начало
		// Создаем новый срез: сначала все values, потом старые значения
		newItems := make([]interface{}, len(values)+len(list.Items))

		// Копируем новые значения в начало
		copy(newItems, values)

		// Копируем старые значения после новых
		copy(newItems[len(values):], list.Items)

		// Присваиваем новый срез
		list.Items = newItems

		s.updateStats(SetOp)
		return nil
	}

	// 5. Случай 2 - Создаем новый список
	s.data[key] = NewList()
	s.types[key] = TypeList
	list := s.data[key].(*ListValue)
	list.Items = append(list.Items, values...)

	// 6. Обновляем статистику
	s.updateStats(SetOp)
	return nil
}

// RPush - добавляем один или несколько элементов в КОНЕЦ списка
// Если ключ не существует, создает новый список
// Если ключ существует, но не список - возвращает ошибку
func (s *Store) RPush(key string, values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	// Блокируем для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// СЛУЧАЙ 1 - Ключ уже существует
	if val, exists := s.data[key]; exists {

		// Проверяем тип
		if s.types[key] != TypeList {
			return fmt.Errorf("key %s exists but is not a list (type:%s)",
				key, s.types[key].String())
		}

		// 3. Добавляем к существующему списку
		list := val.(*ListValue)
		list.Items = append(list.Items, values...)

		s.updateStats(SetOp)
		return nil
	}

	// 4. 2 Случай - Создаем новый список
	s.data[key] = NewList()
	s.types[key] = TypeList
	list := s.data[key].(*ListValue)
	list.Items = append(list.Items, values...)

	s.updateStats(SetOp)
	return nil
}

// RPop - удаляет и возвращает последний элемент
func (s *Store) RPop(key string) (interface{}, error) {
	// 1. Заблокировать мьютекс для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Проверить существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key %s does not exist", key)
	}

	// 3. Проверяем, что это список
	if s.types[key] != TypeList {
		return nil, fmt.Errorf("key %s exists but not is a list (type:%s)",
			key, s.types[key].String())
	}

	// 3. Проверяем, что список не пустой
	list := val.(*ListValue)
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("list %s is empty", key)
	}

	// 4. Поиск индекса последнего элемента
	lastIndex := len(list.Items) - 1

	// 5. Берем последний элемент
	item := list.Items[lastIndex]

	// 6. Создаем новый срез без последнего элемента
	list.Items = list.Items[:lastIndex]

	// 7. Если список стал пустым - удаляем ключ
	if len(list.Items) == 0 {
		delete(s.data, key)
		delete(s.types, key)
	}

	// 8. Обновляем статистику и возвращаем результат
	s.updateStats(GetOp)
	return item, nil
}

// LPop - удаляет и возвращает первый элемент
func (s *Store) LPop(key string) (interface{}, error) {
	// 1. Заблокировать мьютекс для записи
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Проверить, существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key %s does not exist", key)
	}

	// 3. Проверить, что это список
	if s.types[key] != TypeList {
		return nil, fmt.Errorf("key %s exists but not is a list (type:%s)",
			key, s.types[key].String())
	}

	// 4. Получаем список
	list := val.(*ListValue)

	// 5. Проверяем, что не список не пустой
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("list %s is empty", key)
	}

	// 6. Берем первый элемент
	item := list.Items[0]

	// 7. Создаем новый массив без первого элемента
	list.Items = list.Items[1:]

	// 8. Если список стал пустым - удаляем ключ
	if len(list.Items) == 0 {
		delete(s.data, key)
		delete(s.types, key)
	}

	// 9. Обновляет статистику и возвращаем результат
	s.updateStats(GetOp)
	return item, nil
}

// LLen - возвращаем длину списка
func (s *Store) LLen(key string) (int, error) {
	// 1. Блокируем мьютекс для чтения
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 2. Проверяем, существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return 0, nil
	}

	// 3. Проверяем тип
	if s.types[key] != TypeList {
		return 0, fmt.Errorf("key %s is exists but is not a list (type:%s)",
			key, s.types[key].String())
	}

	// 4. Получаем список и возвращаем его длину
	list := val.(*ListValue)
	return len(list.Items), nil
}

// LIndex - возвращает элемент по индексу (не удаляя)
func (s *Store) LIndex(key string, index int) (interface{}, error) {
	// 1. Блокируем мьютекс для чтения
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 2. Проверяем, существует ли ключ
	val, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key %s doest not exist", key)
	}

	// 3. Проверяем тип
	if s.types[key] != TypeList {
		return nil, fmt.Errorf("key %s is exists but is not a list (type:%s)",
			key, s.types[key].String())
	}

	// 4. Получаем список
	list := val.(*ListValue)

	// 5. Обрабатываем отрицательный индекс
	if index < 0 {
		index = len(list.Items) + index
	}

	// 6. Проверяем границы
	if index < 0 || index >= len(list.Items) {
		return nil, fmt.Errorf("index %d out of range (list size: %d)",
			index, len(list.Items))
	}

	return list.Items[index], nil
}

// LRange - возвращает диапазон элементов списка от start до stop
// start - начальный идекс (включительно)
// stop - конечный индекс (включительно)
// Поддерживает отрицательные индексы (-1 = последний, -2 = предпоследний)
func (s *Store) LRange(key string, start, stop int) (interface{}, error) {
	// 1. Блокируем для чтения
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 2. Проверяем, существует ли ключ
	val, exist := s.data[key]
	if !exist {
		return []interface{}{}, nil
	}

	// 3. Проверяем, что это список
	if s.types[key] != TypeList {
		return nil, fmt.Errorf("key %s is exists but is not a list (type:%s)",
			key, s.types[key].String())
	}

	// 4. Получаем список
	list := val.(*ListValue)

	// 5. Получаем длину списка
	length := len(list.Items)
	if length == 0 {
		return []interface{}{}, nil
	}

	// 6. Нормализуем индексы (обрабатываем отрицательные)
	if start < 0 {
		start = length + start
	}

	if stop < 0 {
		stop = length + stop
	}

	// 7. Проверяем границы
	if start < 0 {
		start = 0
	}

	if stop >= length {
		stop = length - 1
	}

	if start > stop || start >= length {
		return []interface{}{}, nil
	}

	// 8. Создаем срез результата нужного размера
	resultSize := stop - start + 1
	result := make([]interface{}, resultSize)

	// 9. Копируем элементы
	copy(result, list.Items[start:stop+1])

	// 10 Обновляем статистику
	s.updateStats(GetOp)
	return result, nil
}
