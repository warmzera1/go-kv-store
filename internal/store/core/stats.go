package core

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
func (s *Store) UpdateStats(op int) {
	s.StatsMu.Lock()
	defer s.StatsMu.Unlock()

	switch op {
	case SetOp:
		s.Stats.SetCount++
	case GetOp:
		s.Stats.GetCount++
	case DelOp:
		s.Stats.DelCount++
	}
}

func (s *Store) GetStats() Stats {
	// 1. Блокируем данные для чтения
	s.StatsMu.Lock()
	defer s.StatsMu.Unlock()

	// 2. Возвращаем копию, чтобы нельзя было изменить оригинал
	return Stats{
		SetCount: s.Stats.SetCount,
		GetCount: s.Stats.GetCount,
		DelCount: s.Stats.DelCount,
	}
}
