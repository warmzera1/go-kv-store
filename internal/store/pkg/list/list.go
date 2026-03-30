package list

import (
	"fmt"
	"slices"

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

	// 1.1 Проверяем, истек ли ключ, если да - удаляем
	s.ttl.IsExpiredAndClean(key)
	delete(s.store.Expiry, key)

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
		slices.Reverse(values)
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
	slices.Reverse(values)
	list.Items = append(values, list.Items...)

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

	s.ttl.IsExpiredAndClean(key)
	delete(s.store.Expiry, key)

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

	// 6. Проверяем, что список не пустой
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
