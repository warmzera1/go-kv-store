package store

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
