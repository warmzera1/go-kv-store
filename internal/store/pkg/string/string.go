package str

import (
	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

// StringStore - обертка над store.Store , реализующая InterfaceStore
type StringStore struct {
	store *core.Store
	ttl   ttl.TTLInterface
}

// New создает новый StringStore
func New(s *core.Store, t ttl.TTLInterface) *StringStore {
	return &StringStore{
		store: s,
		ttl:   t,
	}
}

// Set - добавить элемент
func (s *StringStore) Set(key, value string) {
	// 1. Блокируем запись в data
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 3. Сохраняем значение
	s.store.Data[key] = core.StringValue(value)
	s.store.Types[key] = core.TypeString

	// 4. Обновляем статистику
	s.store.UpdateStats(core.SetOp)
}

// Get - получить элемент
func (s *StringStore) Get(key string) (string, bool) {
	// 1. Блокируем данные для чтения
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 2. Проверяем, истек ли ключ, если да - удаляем
	if s.ttl.IsExpiredAndClean(key) {
		return "", false
	}

	// 3. Проверяем, существует ли ключ и проверяем тип
	val, exists := s.store.Data[key]
	if !exists || s.store.Types[key] != core.TypeString {
		return "", false
	}

	// 4. Обновляем статистику
	s.store.UpdateStats(core.GetOp)

	// 5. Возвращаем результат
	return string(val.(core.StringValue)), true
}

func (s *StringStore) Delete(key string) {
	// 1. Блокируем данные для чтения
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	// 2. Удаляем
	delete(s.store.Data, key)
	delete(s.store.Types, key)

	// 3. Обновляем статистику
	s.store.UpdateStats(core.DelOp)
}
