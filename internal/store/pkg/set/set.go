package set

import (
	"fmt"

	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

type SetStore struct {
	store *core.Store
	ttl   ttl.TTLInterface
}

func New(s *core.Store, t ttl.TTLInterface) *SetStore {
	return &SetStore{
		store: s,
		ttl:   t,
	}
}

// SAdd - добавляет один или несколько элементов в множество
// Если элемент существует - игнорируется
// Если ключ не сущетсвуте - создает новое множество
// Если ключ существует, но не множество - возвращает ошибку
func (s *SetStore) SAdd(key string, values ...interface{}) (int, error) {
	// 1. Проверяем, есть ли что добавить
	if len(values) == 0 {
		return 0, nil
	}

	// 2. Блокируем хранилище для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 3. СЛУЧАЙ 1 - Проверяем, существует ли ключ
	if val, exists := s.store.Data[key]; exists {
		// 4. Проверяем, что тип - множество
		if s.store.Types[key] != core.TypeSet {
			return 0, fmt.Errorf("key %s exists but is not a set (type: %s)",
				key, s.store.Types[key].String())
		}

		// 5. Получаем множество
		set := val.(*core.SetValue)

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
			s.store.UpdateStats(core.SetOp)
		}

		return added, nil
	}

	// 8. СЛУЧАЙ 2 - Ключа нет - создаем новое множество
	s.store.Data[key] = core.NewSet()
	s.store.Types[key] = core.TypeSet
	set := s.store.Data[key].(*core.SetValue)

	added := 0
	for _, v := range values {
		if _, exists := set.Items[v]; !exists {
			set.Items[v] = struct{}{}
			added++
		}
	}

	if added > 0 {
		s.store.UpdateStats(core.SetOp)
	}

	return added, nil
}

// SRem - удаляет один или несколько элементов из множества
// Если ключа нет - возвращаем nil
// Если ключ есть, но не множество - ошибка
// Если множество после удаления, становится пустым - удаляем ключ
func (s *SetStore) SRem(key string, values ...interface{}) (int, error) {
	// 1. Проверка на пустой ввод
	if len(values) == 0 {
		return 0, nil
	}

	// 2. Блокируем операцию для записи
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 3. СЛУЧАЙ 1 - Проверка, существует ли ключ
	if val, exists := s.store.Data[key]; exists {
		if s.store.Types[key] != core.TypeSet {
			return 0, fmt.Errorf("key %s is exists but is not a set (type:%s)",
				key, s.store.Types[key].String())
		}

		// 4. Получаем множество
		set := val.(*core.SetValue)

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
			delete(s.store.Data, key)
			delete(s.store.Types, key)
		}

		// 9. Обновляем статистику
		if removed > 0 {
			s.store.UpdateStats(core.DelOp)
		}

		return removed, nil
	}

	return 0, nil
}

// SIsMember - проверяет, есть ли элемент в множестве
func (s *SetStore) SIsMember(key string, value interface{}) (bool, error) {
	// 1 Блокируем на чтение
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 2. Проверяем истек ли ключ, если да - удаляем
	if s.ttl.IsExpiredAndClean(key) {
		return false, nil
	}

	// 3. Проверка, существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return false, nil
	}

	if s.store.Types[key] != core.TypeSet {
		return false, fmt.Errorf("key %s exists but is not a set (type:%s)",
			key, s.store.Types[key].String())
	}

	// 4. Получаем множество
	set := val.(*core.SetValue)

	// 5. Проверяем наличие элемента
	_, found := set.Items[value]

	// 6. Обновляем статистику
	s.store.UpdateStats(core.GetOp)

	// 7. Возвращаем результат
	return found, nil
}

// SMembers - возвращает все элементы множества
func (s *SetStore) SMembers(key string) ([]interface{}, error) {
	// 1. Блокируем для чтения
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 2. Проверка, истек ли ключ, если да - удаляем
	if s.ttl.IsExpiredAndClean(key) {
		return nil, nil
	}

	// 3. Проверка существования ключа и типа, если ключ существует
	val, exists := s.store.Data[key]
	if !exists {
		return []interface{}{}, nil
	}

	if s.store.Types[key] != core.TypeSet {
		return nil, fmt.Errorf("key %s exists but not is a set (type:%s)",
			key, s.store.Types[key].String())
	}

	// 3. Получение множества
	set := val.(*core.SetValue)

	// 4. Создание среза результата
	result := make([]interface{}, 0, len(set.Items))

	// 5. Копирование элементов
	for v := range set.Items {
		result = append(result, v)
	}

	// 6. Обновление статистики
	s.store.UpdateStats(core.GetOp)

	return result, nil
}

// SCard - возвращает количество элементов в множестве
func (s *SetStore) SCard(key string) (int, error) {
	// 1. Блокируем для чтения
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	// 2. Проверяем, есть ли ключ (без удаления)
	if s.ttl.IsExpired(key) {
		return 0, nil
	}

	// 3. Проверяем, существует ли ключ
	val, exists := s.store.Data[key]
	if !exists {
		return 0, nil
	}

	// 3. Проверяем тип
	if s.store.Types[key] != core.TypeSet {
		return 0, fmt.Errorf("key %s exists but is not a set (type:%s)",
			key, s.store.Types[key].String())
	}

	// 4. Получение множества
	set := val.(*core.SetValue)

	return len(set.Items), nil
}
