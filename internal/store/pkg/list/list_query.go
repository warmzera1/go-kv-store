package list

import (
	"fmt"

	"github.com/warmzera1/kv-store/internal/store/core"
)

// LLen - возвращаем длину списка
func (s *ListStore) LLen(key string) (int, error) {
	// 1. Блокируем мьютекс для чтения
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	// 2. Проверяем, существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return 0, nil
	}

	// 3. Проверяем тип
	if s.store.Types[key] != core.TypeList {
		return 0, fmt.Errorf("key %s is exists but is not a list (type:%s)",
			key, s.store.Types[key].String())
	}

	// 4. Получаем список и возвращаем его длину
	list := val.(*core.ListValue)
	return len(list.Items), nil
}

// LIndex - возвращает элемент по индексу (не удаляя)
func (s *ListStore) LIndex(key string, index int) (interface{}, error) {
	// 1. Блокируем мьютекс для чтения
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	// 2. Проверяем, истек ли ключ (без удаления)
	if s.ttl.IsExpired(key) {
		return 0, nil
	}

	// 3. Проверяем, существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return nil, fmt.Errorf("key %s doest not exist", key)
	}

	// 4. Проверяем тип
	if s.store.Types[key] != core.TypeList {
		return nil, fmt.Errorf("key %s is exists but is not a list (type:%s)",
			key, s.store.Types[key].String())
	}

	// 5. Получаем список
	list := val.(*core.ListValue)

	// 6. Обрабатываем отрицательный индекс
	if index < 0 {
		index = len(list.Items) + index
	}

	// 7. Проверяем границы
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
func (s *ListStore) LRange(key string, start, stop int) ([]interface{}, error) {
	// 1. Блокируем для чтения
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 2. Проверяем, истек ли ключ, если да - удаляем
	if s.ttl.IsExpiredAndClean(key) {
		return nil, nil
	}

	// 3. Проверяем, существует ли ключ
	val, exist := s.store.Data[key]
	if !exist {
		return []interface{}{}, nil
	}

	// 4. Проверяем, что это список
	if s.store.Types[key] != core.TypeList {
		return nil, fmt.Errorf("key %s is exists but is not a list (type:%s)",
			key, s.store.Types[key].String())
	}

	// 5. Получаем список
	list := val.(*core.ListValue)

	// 6. Получаем длину списка
	length := len(list.Items)
	if length == 0 {
		return []interface{}{}, nil
	}

	// 7. Нормализуем индексы (обрабатываем отрицательные)
	if start < 0 {
		start = length + start
	}

	if stop < 0 {
		stop = length + stop
	}

	// 8. Проверяем границы
	if start < 0 {
		start = 0
	}

	if stop >= length {
		stop = length - 1
	}

	if start > stop || start >= length {
		return []interface{}{}, nil
	}

	// 9. Создаем срез результата нужного размера
	resultSize := stop - start + 1
	result := make([]interface{}, resultSize)

	// 10. Копируем элементы
	copy(result, list.Items[start:stop+1])

	// 11 Обновляем статистику
	s.store.UpdateStats(core.GetOp)
	return result, nil
}
