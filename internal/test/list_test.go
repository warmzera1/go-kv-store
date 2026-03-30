package test

import (
	"testing"
	"time"

	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

func TestLPushSingle(t *testing.T) {
	_, listStore := NewTestListStore()

	err := listStore.LPush("key", "value")
	if err != nil {
		t.Fatalf("LPush failed: '%v'", err)
	}

	length, err := listStore.LLen("key")
	if err != nil {
		t.Fatalf("LLen failed: '%v'", err)
	}
	if length != 1 {
		t.Errorf("Expected length 1, got '%d'", length)
	}

	val, err := listStore.LIndex("key", 0)
	if err != nil {
		t.Fatalf("LIndex failed: '%v'", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got '%v'", val)
	}
}

func TestLPushMultiple(t *testing.T) {
	_, listStore := NewTestListStore()

	err := listStore.LPush("key", "a", "b", "c")
	if err != nil {
		t.Fatalf("LPush failed: '%v'", err)
	}

	length, err := listStore.LLen("key")
	if err != nil {
		t.Fatalf("LLen failed: '%v'", err)
	}
	if length != 3 {
		t.Errorf("Expected length 3, got '%d'", length)
	}

	expected := []string{"c", "b", "a"}
	for i, exp := range expected {
		val, err := listStore.LIndex("key", i)
		if err != nil {
			t.Fatalf("LIndex at %d failed: '%v'", i, err)
		}
		if val != exp {
			t.Errorf("At index %d: expected '%s', got '%v'", i, exp, val)
		}
	}
}

func TestLPushOnExisting(t *testing.T) {
	_, listStore := NewTestListStore()

	err := listStore.LPush("key", "a", "b")
	if err != nil {
		t.Errorf("LPush failed: '%v'", err)
	}

	err = listStore.LPush("key", "c", "d")
	if err != nil {
		t.Errorf("LPush failed: '%v'", err)
	}

	expected := []string{"d", "c", "b", "a"}
	for i, exp := range expected {
		val, err := listStore.LIndex("key", i)
		if err != nil {
			t.Fatalf("LIndex at %d failed: '%v'", i, err)
		}
		if val != exp {
			t.Errorf("At index %d: expected '%s', got %v", i, exp, val)
		}
	}
}

func TestLPushWithTTL(t *testing.T) {
	s, listStore := NewTestListStore()
	ttlStore := ttl.New(s)

	ttlStore.StartTTLCleaner(100 * time.Millisecond)

	listStore.LPush("key", "a", "b")
	ttlStore.Expire("key", 1)

	time.Sleep(1100 * time.Millisecond)

	length, err := listStore.LLen("key")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 0 {
		t.Errorf("Expected length 0 after expiration, got '%d'", length)
	}
}

func TestLPushWrongType(t *testing.T) {
	s, listStore := NewTestListStore()

	key := "wrong_type_key"
	s.Mu.Lock()
	s.Data[key] = &core.ListValue{
		Items: []interface{}{"1", "2"},
	}
	s.Types[key] = core.TypeSet
	s.Mu.Unlock()

	err := listStore.LPush(key, "a", "b")
	if err == nil {
		t.Error("Expected error for wrong type (String), but LPush succeeded")
	}
}

func TestRPushSingle(t *testing.T) {
	_, listStore := NewTestListStore()

	err := listStore.RPush("key", "value")
	if err != nil {
		t.Fatalf("RPush failed: '%v'", err)
	}

	length, err := listStore.LLen("key")
	if err != nil {
		t.Fatalf("LLen failed: '%v'", err)
	}
	if length != 1 {
		t.Errorf("Expected length 1, got '%d'", length)
	}

	val, err := listStore.LIndex("key", 0)
	if err != nil {
		t.Fatalf("LIndex failed: '%v'", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got '%v'", val)
	}
}

func TestRPushMultiple(t *testing.T) {
	_, listStore := NewTestListStore()

	err := listStore.RPush("key", "a", "b", "c")
	if err != nil {
		t.Fatalf("RPush failed: '%v'", err)
	}

	length, err := listStore.LLen("key")
	if err != nil {
		t.Fatalf("LLen failed: '%v'", err)
	}
	if length != 3 {
		t.Errorf("Expected length 3, got '%d'", length)
	}

	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		val, err := listStore.LIndex("key", i)
		if err != nil {
			t.Fatalf("LIndex at %d failed: '%v'", i, err)
		}
		if val != exp {
			t.Errorf("At index %d: expected '%s', got '%v'", i, exp, val)
		}
	}
}

func TestRPushOnExisting(t *testing.T) {
	_, listStore := NewTestListStore()

	err := listStore.RPush("key", "a", "b")
	if err != nil {
		t.Errorf("LPush failed: '%v'", err)
	}

	err = listStore.RPush("key", "c", "d")
	if err != nil {
		t.Errorf("LPush failed: '%v'", err)
	}

	expected := []string{"a", "b", "c", "d"}
	for i, exp := range expected {
		val, err := listStore.LIndex("key", i)
		if err != nil {
			t.Fatalf("LIndex at %d failed: '%v'", i, err)
		}
		if val != exp {
			t.Errorf("At index %d: expected '%s', got %v", i, exp, val)
		}
	}
}

func TestRPushWrongType(t *testing.T) {
	s, listStore := NewTestListStore()

	key := "wrong_type_key"
	s.Mu.Lock()
	s.Data[key] = &core.ListValue{
		Items: []interface{}{"1", "2"},
	}
	s.Types[key] = core.TypeSet
	s.Mu.Unlock()

	err := listStore.RPush(key, "a", "b")
	if err == nil {
		t.Error("Expected error for wrong type (String), but RPush succeeded")
	}
}

func TestRPushWithTTL(t *testing.T) {
	s, listStore := NewTestListStore()
	ttlStore := ttl.New(s)

	ttlStore.StartTTLCleaner(100 * time.Millisecond)

	listStore.RPush("key", "a", "b")
	ttlStore.Expire("key", 1)

	time.Sleep(1100 * time.Millisecond)

	length, err := listStore.LLen("key")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 0 {
		t.Errorf("Expected length 0 after expiration, got '%d'", length)
	}
}

func TestLPopSingle(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.LPush("key", "a", "b")
	item, err := listStore.LPop("key")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if item != "b" {
		t.Errorf("Expected item = %v after deleted", item)
	}

	length, _ := listStore.LLen("key")
	if length != 1 {
		t.Errorf("Expected length 1, got %d", length)
	}
}

func TestLPopEmpty(t *testing.T) {
	_, listStore := NewTestListStore()

	item, err := listStore.LPop("key")
	if err == nil {
		t.Error("Expected error for empty list, got nil")
	}
	if item != nil {
		t.Errorf("Expected nil item, got '%v'", item)
	}
}

func TestLPopUntilEmpty(t *testing.T) {
	_, listStore := NewTestListStore()

	err := listStore.LPush("key", "a", "b")
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}

	item, err := listStore.LPop("key")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if item != "b" {
		t.Errorf("Expected item = %v after deleted", item)
	}

	item, err = listStore.LPop("key")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if item != "a" {
		t.Errorf("Expected item = %v after deleted", item)
	}

	length, _ := listStore.LLen("key")
	if length != 0 {
		t.Errorf("Expected length 0, got %d", length)
	}
}

func TestLPopWrongType(t *testing.T) {
	s, listStore := NewTestListStore()

	key := "wrong_type_key"
	s.Mu.Lock()
	s.Data[key] = &core.ListValue{
		Items: []interface{}{"1", "2"},
	}
	s.Types[key] = core.TypeSet
	s.Mu.Unlock()

	item, err := listStore.LPop(key)
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
	if item != nil {
		t.Error("Expected error for wrong type (Set)")
	}
}

func TestRPopSingle(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.LPush("key", "a", "b")
	item, err := listStore.RPop("key")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if item != "a" {
		t.Errorf("Expected item = %v after deleted", item)
	}

	length, _ := listStore.LLen("key")
	if length != 1 {
		t.Errorf("Expected length 1, got %d", length)
	}
}

func TestRPopEmpty(t *testing.T) {
	_, listStore := NewTestListStore()

	item, err := listStore.RPop("key")
	if err == nil {
		t.Error("Expected error for empty list, got nil")
	}
	if item != nil {
		t.Errorf("Expected nil item, got '%v'", item)
	}
}

func TestRPopUntilEmpty(t *testing.T) {
	_, listStore := NewTestListStore()

	err := listStore.LPush("key", "a", "b")
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}

	item, err := listStore.RPop("key")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if item != "a" {
		t.Errorf("Expected item = %v after deleted", item)
	}

	item, err = listStore.RPop("key")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if item != "b" {
		t.Errorf("Expected item = %v after deleted", item)
	}

	length, _ := listStore.LLen("key")
	if length != 0 {
		t.Errorf("Expected length 0, got %d", length)
	}
}

func TestRPopWrongType(t *testing.T) {
	s, listStore := NewTestListStore()

	key := "wrong_type_key"
	s.Mu.Lock()
	s.Data[key] = &core.ListValue{
		Items: []interface{}{"1", "2"},
	}
	s.Types[key] = core.TypeSet
	s.Mu.Unlock()

	item, err := listStore.RPop(key)
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
	if item != nil {
		t.Error("Expected error for wrong type (Set)")
	}
}

func TestLLenBasic(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.RPush("key", "a", "b", "c")
	length, err := listStore.LLen("key")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}
}

func TestLLenEmpty(t *testing.T) {
	_, listStore := NewTestListStore()

	length, err := listStore.LLen("key")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 0 {
		t.Errorf("Expected length 0, got %d", length)
	}

	listStore.RPush("empty")
	length, err = listStore.LLen("empty")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 0 {
		t.Errorf("Expected length 0, got %d", length)
	}
}

func TestLLenWrongType(t *testing.T) {
	s, listStore := NewTestListStore()

	key := "wrong_type_key"

	s.Mu.Lock()
	s.Data[key] = "string"
	s.Types[key] = core.TypeString
	s.Mu.Unlock()

	_, err := listStore.LLen(key)
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestLIndexPositive(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.RPush("key", "a", "b", "c")
	item, err := listStore.LIndex("key", 2)
	if err != nil {
		t.Fatalf("LIndex failed: %v", err)
	}
	if item != "c" {
		t.Errorf("Expected item = '%v'", item)
	}

	item, err = listStore.LIndex("key", 0)
	if err != nil {
		t.Fatalf("LIndex failed: %v", err)
	}
	if item != "a" {
		t.Errorf("Expected item = '%v'", item)
	}
}

func TestLIndexNegative(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.RPush("key", "a", "b", "c")
	item, err := listStore.LIndex("key", -3)
	if err != nil {
		t.Fatalf("LIndex failed: %v", err)
	}
	if item != "a" {
		t.Errorf("Expected item = '%v'", item)
	}

	item, err = listStore.LIndex("key", -1)
	if err != nil {
		t.Fatalf("LIndex failed: %v", err)
	}
	if item != "c" {
		t.Errorf("Expected item = '%v'", item)
	}
}

func TestLIndexOutOfRange(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.LPush("key", "a", "b", "c")

	_, err := listStore.LIndex("key", 5)
	if err == nil {
		t.Error("Expected error for index 5 (out of range), got nil")
	}

	_, err = listStore.LIndex("key", -5)
	if err == nil {
		t.Error("Expected error for index -5 (out of range), got nil")
	}
}

func TestLIndexEmpty(t *testing.T) {
	_, listStore := NewTestListStore()

	_, err := listStore.LIndex("noneexistent", 0)
	if err == nil {
		t.Error("Expected error for non-existent key, got nil")
	}

	listStore.LPush("empty")
	_, err = listStore.LIndex("empty", 0)
	if err == nil {
		t.Error("Expected error for empty list, got nil")
	}
}

func TestLIndexWrongType(t *testing.T) {
	s, listStore := NewTestListStore()

	key := "wrong_type_key"
	s.Mu.Lock()
	s.Data[key] = "string"
	s.Types[key] = core.TypeString
	s.Mu.Unlock()

	_, err := listStore.LIndex(key, 0)
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestLRangeAll(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.RPush("key", "a", "b", "c")

	items, err := listStore.LRange("key", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("Expected length 3, got %d", len(items))
	}

	expected := []interface{}{"a", "b", "c"}
	for i, exp := range expected {
		if items[i] != exp {
			t.Errorf("At index %d: expected '%v', got '%v'", i, exp, items[i])
		}
	}
}

func TestLRangeFirstTwo(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.RPush("key", "a", "b", "c")

	items, err := listStore.LRange("key", 0, 1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected length 2, got %d", len(items))
	}

	expected := []interface{}{"a", "b"}
	for i, exp := range expected {
		if items[i] != exp {
			t.Errorf("At index %d: expected '%v', got '%v'", i, exp, items[i])
		}
	}
}

func TestLRangeLastTwo(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.RPush("key", "a", "b", "c")

	items, err := listStore.LRange("key", -2, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected length 2, got %d", len(items))
	}

	expected := []interface{}{"b", "c"}
	for i, exp := range expected {
		if items[i] != exp {
			t.Errorf("At index %d: expected '%v', got '%v'", i, exp, items[i])
		}
	}
}

func TestLRangeMiddle(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.RPush("key", "a", "b", "c", "d", "e")

	items, err := listStore.LRange("key", 1, 3)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("Expected length 3, got %d", len(items))
	}

	expected := []interface{}{"b", "c", "d"}
	for i, exp := range expected {
		if items[i] != exp {
			t.Errorf("At index %d: expected '%v', got '%v'", i, exp, items[i])
		}
	}
}

func TestLRangeInvalid(t *testing.T) {
	_, listStore := NewTestListStore()

	listStore.RPush("key", "a", "b", "c")

	items, err := listStore.LRange("key", 3, 1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected length 0, got %d", len(items))
	}
}

func TestLRangeEmpty(t *testing.T) {
	_, listStore := NewTestListStore()

	items, err := listStore.LRange("key", 1, 3)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected length 0, got %d", len(items))
	}
}

func TestLRangeWrongType(t *testing.T) {
	s, listStore := NewTestListStore()

	key := "wrong_type_key"
	s.Mu.Lock()
	s.Data[key] = "string"
	s.Types[key] = core.TypeString
	s.Mu.Unlock()

	_, err := listStore.LRange(key, 1, 3)
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}
