package test

import (
	"testing"

	"github.com/warmzera1/kv-store/internal/store/core"
)

func TestHSetNewField(t *testing.T) {
	_, hashStore := NewTestHashStore()

	err := hashStore.HSet("key", "field1", "value1")
	if err != nil {
		t.Fatalf("Failed HSet: %v", err)
	}

	item, err := hashStore.HGet("key", "field1")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "value1" {
		t.Errorf("Expected 'value1', got %v", item)
	}

	err = hashStore.HSet("key", "field2", "value2")
	if err != nil {
		t.Fatalf("Failed HSet: %v", err)
	}

	item, err = hashStore.HGet("key", "field2")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "value2" {
		t.Errorf("Expected 'value2', got %v", item)
	}

	members, err := hashStore.HGetAll("key")
	if err != nil {
		t.Fatalf("Failed HGetAll: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
}

func TestHSetUpdateField(t *testing.T) {
	_, hashStore := NewTestHashStore()

	err := hashStore.HSet("key", "field", "value1")
	if err != nil {
		t.Fatalf("Failed HSet: %v", err)
	}

	item, err := hashStore.HGet("key", "field")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "value1" {
		t.Errorf("Expected 'value1', got %v", item)
	}

	err = hashStore.HSet("key", "field", "value2")
	if err != nil {
		t.Fatalf("Failed HSet: %v", err)
	}

	item, err = hashStore.HGet("key", "field")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "value2" {
		t.Errorf("Expected 'value2', got %v", item)
	}

	members, err := hashStore.HGetAll("key")
	if err != nil {
		t.Fatalf("Failed HGetAll: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("Expected 1 members, got %d", len(members))
	}
}

func TestHGetBasic(t *testing.T) {
	_, hashStore := NewTestHashStore()

	hashStore.HSet("key", "field1", "value1")
	hashStore.HSet("key", "field2", "value2")

	item, err := hashStore.HGet("key", "field1")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "value1" {
		t.Errorf("Expected 'value1', got %v", item)
	}

	item, err = hashStore.HGet("key", "field2")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "value2" {
		t.Errorf("Expected 'value2', got %v", item)
	}
}

func TestHGetNonExist(t *testing.T) {
	_, hashStore := NewTestHashStore()

	hashStore.HSet("key", "field", "value")

	val, err := hashStore.HGet("key", "noneexist")
	if err != nil {
		t.Fatalf("HGet returned error: %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil for non-existent filed, got %v", val)
	}

	val, err = hashStore.HGet("noneexist", "field")
	if err != nil {
		t.Fatalf("HGet returned error: %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil for non-existent key, got %v", val)
	}

	val, err = hashStore.HGet("noneexist", "noneexist")
	if err != nil {
		t.Fatalf("HGet returned error: %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil for non-existend key and filed, got %v", val)
	}
}

func TestHGetAll(t *testing.T) {
	_, hashStore := NewTestHashStore()

	err := hashStore.HSet("key", "name", "Alice")
	if err != nil {
		t.Fatalf("Failed HSet: %v", err)
	}

	err = hashStore.HSet("key", "age", "10")
	if err != nil {
		t.Fatalf("Failed HSet: %v", err)
	}

	members, err := hashStore.HGetAll("key")
	if err != nil {
		t.Fatalf("Failed HGetAll: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}

	item, err := hashStore.HGet("key", "name")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "Alice" {
		t.Errorf("Expected 'Alice', got %v", item)
	}

	item, err = hashStore.HGet("key", "age")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "10" {
		t.Errorf("Expected '10', got %v", item)
	}

	members, err = hashStore.HGetAll("noneexist")
	if err != nil {
		t.Fatalf("HGetAll returned error: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("Expected empty map for non-existent key, got %d items", len(members))
	}
}

func TestHDelSingle(t *testing.T) {
	_, hashStore := NewTestHashStore()

	hashStore.HSet("key", "field1", "value1")
	hashStore.HSet("key", "field2", "value2")

	err := hashStore.HDel("key", "field1")
	if err != nil {
		t.Fatalf("Failed HDel: %v", err)
	}

	val, err := hashStore.HGet("key", "field1")
	if err != nil {
		t.Fatalf("HGet returned error: %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil for deleted field, got %d", val)
	}

	item, err := hashStore.HGet("key", "field2")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "value2" {
		t.Errorf("Expected 'value2', got %v", item)
	}

	members, err := hashStore.HGetAll("key")
	if err != nil {
		t.Fatalf("Failed HGetAll: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("Expected 1 members, got %d", len(members))
	}
}

func TestHDelMultiple(t *testing.T) {
	_, hashStore := NewTestHashStore()

	hashStore.HSet("key", "field1", "v1")
	hashStore.HSet("key", "field2", "v2")
	hashStore.HSet("key", "field3", "v3")

	err := hashStore.HDel("key", "field1", "field2")
	if err != nil {
		t.Fatalf("Failed HDel: %v", err)
	}

	members, err := hashStore.HGetAll("key")
	if err != nil {
		t.Fatalf("Failed HGetAll: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("Expected 1 members, got %d", len(members))
	}

	item, err := hashStore.HGet("key", "field3")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != "v3" {
		t.Errorf("Expected 'v3', got %v", item)
	}
}

func TestHDelEmptyHash(t *testing.T) {
	_, hashStore := NewTestHashStore()

	hashStore.HSet("key", "field", "v")
	err := hashStore.HDel("key", "field")
	if err != nil {
		t.Fatalf("Failed HDel: %v", err)
	}

	members, err := hashStore.HGetAll("key")
	if err != nil {
		t.Fatalf("Failed HGetAll: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("Expected 0 members, got %d", len(members))
	}

	item, err := hashStore.HGet("key", "field")
	if err != nil {
		t.Fatalf("Failed HGet: %v", err)
	}
	if item != nil {
		t.Errorf("Expected '', got %v", item)
	}

	hashStore.HSet("key", "field", "v")
	err = hashStore.HDel("key", "field")
	if err != nil {
		t.Fatalf("Failed HDel: %v", err)
	}
}

func TestHSetWrongType(t *testing.T) {
	s, hashStore := NewTestHashStore()

	key := "wrong_type_key"

	s.Mu.Lock()
	s.Data[key] = "string"
	s.Types[key] = core.TypeString
	s.Mu.Unlock()

	err := hashStore.HSet(key, "field", "value1")
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}

	val, err := hashStore.HGet(key, "field")
	if err == nil {
		t.Fatalf("HGet returned error: %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}
