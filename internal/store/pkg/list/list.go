package list

import (
	"fmt"

	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

type ListStore struct {
	store *core.Store
	ttl   ttl.TTLInterface
}

func New(s *core.Store, t ttl.TTLInterface) *ListStore {
	return &ListStore{
		store: s,
		ttl:   t,
	}
}

// LPush - добавляет элемент в НАЧАЛО списка
// Если ключ не существует, создает новый список
// Если ключ существует, но не список - возвращает ошибку
func (s *ListStore) LPush(key string, values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	// 1. Блокируем для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 2. Случай 1 - Ключ уже существует
	if val, exists := s.store.Data[key]; exists {
		if s.store.Types[key] != core.TypeList {
			return fmt.Errorf("key %s exists but is not a list (type:%s)",
				key, s.store.Types[key].String())
		}

		// 3. Приводим к типу
		list := val.(*core.ListValue)

		// 4. Добавляем в начало
		// Создаем новый срез: сначала все values, потом старые значения
		newItems := make([]interface{}, len(values)+len(list.Items))

		// Копируем новые значения в начало
		copy(newItems, values)

		// Копируем старые значения после новых
		copy(newItems[len(values):], list.Items)

		// Присваиваем новый срез
		list.Items = newItems

		s.store.UpdateStats(core.SetOp)
		return nil
	}

	// 5. Случай 2 - Создаем новый список
	s.store.Data[key] = core.NewList()
	s.store.Types[key] = core.TypeList
	list := s.store.Data[key].(*core.ListValue)
	list.Items = append(list.Items, values...)

	// 6. Обновляем статистику
	s.store.UpdateStats(core.SetOp)
	return nil
}

// RPush - добавляем один или несколько элементов в КОНЕЦ списка
// Если ключ не существует, создает новый список
// Если ключ существует, но не список - возвращает ошибку
func (s *ListStore) RPush(key string, values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	// Блокируем для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// СЛУЧАЙ 1 - Ключ уже существует
	if val, exists := s.store.Data[key]; exists {

		// Проверяем тип
		if s.store.Types[key] != core.TypeList {
			return fmt.Errorf("key %s exists but is not a list (type:%s)",
				key, s.store.Types[key].String())
		}

		// 3. Добавляем к существующему списку
		list := val.(*core.ListValue)
		list.Items = append(list.Items, values...)

		s.store.UpdateStats(core.SetOp)
		return nil
	}

	// 4. 2 Случай - Создаем новый список
	s.store.Data[key] = core.NewList()
	s.store.Types[key] = core.TypeList
	list := s.store.Data[key].(*core.ListValue)
	list.Items = append(list.Items, values...)

	s.store.UpdateStats(core.SetOp)
	return nil
}

// RPop - удаляет и возвращает последний элемент
func (s *ListStore) RPop(key string) (interface{}, error) {
	// 1. Заблокировать мьютекс для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// Проверяем истечение и удаляем, если нужно
	if s.ttl.IsExpiredAndClean(key) {
		return nil, nil
	}

	// 2. Проверить существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return nil, fmt.Errorf("key %s does not exist", key)
	}

	// 3. Проверяем, что это список
	if s.store.Types[key] != core.TypeList {
		return nil, fmt.Errorf("key %s exists but not is a list (type:%s)",
			key, s.store.Types[key].String())
	}

	// 3. Проверяем, что список не пустой
	list := val.(*core.ListValue)
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
		delete(s.store.Data, key)
		delete(s.store.Types, key)
	}

	// 8. Обновляем статистику и возвращаем результат
	s.store.UpdateStats(core.GetOp)
	return item, nil
}

// LPop - удаляет и возвращает первый элемент
func (s *ListStore) LPop(key string) (interface{}, error) {
	// 1. Заблокировать мьютекс для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 2. Проверяем, истек ли ключ, если да - удаляем
	if s.ttl.IsExpiredAndClean(key) {
		return nil, nil
	}

	// 3. Проверить, существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return nil, fmt.Errorf("key %s does not exist", key)
	}

	// 4. Проверить, что это список
	if s.store.Types[key] != core.TypeList {
		return nil, fmt.Errorf("key %s exists but not is a list (type:%s)",
			key, s.store.Types[key].String())
	}

	// 5. Получаем список
	list := val.(*core.ListValue)

	// 6. Проверяем, что не список не пустой
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("list %s is empty", key)
	}

	// 7. Берем первый элемент
	item := list.Items[0]

	// 8. Создаем новый массив без первого элемента
	list.Items = list.Items[1:]

	// 9. Если список стал пустым - удаляем ключ
	if len(list.Items) == 0 {
		delete(s.store.Data, key)
		delete(s.store.Types, key)
	}

	// 10. Обновляет статистику и возвращаем результат
	s.store.UpdateStats(core.GetOp)
	return item, nil
}

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
