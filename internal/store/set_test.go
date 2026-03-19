package store

import (
	"testing"
)

func TestSet(t *testing.T) {
	t.Run("SAdd and SMembers", func(t *testing.T) {
		s := New()

		// Добавляем с дубликатом
		s.SAdd("tags", "golang", "redis", "golang")
		members, err := s.SMembers("tags")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(members) != 2 {
			t.Errorf("Expected 2 members, got %d", len(members))
		}

		// Проверяем, что оба элемента есть
		found := 0
		for _, v := range members {
			if v == "golang" || v == "redis" {
				found++
			}
		}
		if found != 2 {
			t.Errorf("Expected [golang redis], got %v", members)
		}

	})

	t.Run("SIsMember", func(t *testing.T) {
		s := New()
		s.SAdd("tags", "golang", "redis")

		tests := []struct {
			value     interface{}
			exptected bool
		}{
			{"golang", true},
			{"redis", true},
			{"python", false},
			{42, false},
		}

		for _, tt := range tests {
			result, err := s.SIsMember("tags", tt.value)
			if err != nil {
				t.Errorf("Unexpected error for %v: %v", tt.value, err)
			}
			if result != tt.exptected {
				t.Errorf("SIsMember(%v) = %v, want %v", tt.value, result, tt.exptected)
			}
		}
	})

	t.Run("SCard", func(t *testing.T) {
		s := New()

		// Пустое множество
		count, err := s.SCard("nosuchkey")
		if err != nil {
			t.Errorf("Unexpected error for non-existent key: %v", err)
		}
		if count != 0 {
			t.Errorf("SCard(nosuchkey) = %d, want 0", count)
		}

		// С элементами
		s.SAdd("tags", "a", "b", "c")
		count, err = s.SCard("tags")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("SCard(tags) = %d, want 3", count)
		}
	})

	t.Run("SRem", func(t *testing.T) {
		s := New()
		s.SAdd("tags", "a", "b", "c", "d")

		// Удаляем один элемент
		s.SRem("tags", "c")
		members, _ := s.SMembers("tags")
		if len(members) != 3 {
			t.Errorf("After removing c, expected 3 members, got %v", members)
		}

		// Удаляем несуществующий
		s.SRem("tags", "p")
		members, _ = s.SMembers("tags")
		if len(members) != 3 {
			t.Errorf("After removing non-existent, expected still 3, got %v", members)
		}

		// Удаляем все
		s.SRem("tags", "a", "b", "d")
		count, _ := s.SCard("tags")
		if count != 0 {
			t.Errorf("After removing all, expteceted 0, got %v", count)
		}

		// Ключ должен быть удален
		if _, exists := s.data["tags"]; exists {
			t.Error("Key should be deleted when set becomes empty")
		}
	})
}
