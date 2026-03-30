package test

import (
	"testing"
	"time"

	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

func TestSetBasic(t *testing.T) {
	_, strStore := NewTestStringStore()
	strStore.Set("name", "Alice")

	val, ok := strStore.Get("name")
	if !ok {
		t.Error("Expected key to exists")
	}
	if val != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", val)
	}
}

func TestSetOverwrite(t *testing.T) {
	_, strStore := NewTestStringStore()
	strStore.Set("name", "Alice")
	strStore.Set("name", "Bob")

	val, ok := strStore.Get("name")
	if !ok {
		t.Error("Expected key to exist after overwrite")
	}
	if val != "Bob" {
		t.Errorf("Expected 'Bob', got '%s'", val)
	}
}

func TestSetEmptyString(t *testing.T) {
	_, strStore := NewTestStringStore()
	strStore.Set("empty", "")

	val, ok := strStore.Get("empty")
	if !ok {
		t.Error("Empty string should be stored")
	}
	if val != "" {
		t.Errorf("Expected '', got '%s'", val)
	}
}

func TestSetMultipleKeys(t *testing.T) {
	_, strStore := NewTestStringStore()
	strStore.Set("key1", "value1")
	strStore.Set("key2", "value2")

	val, ok := strStore.Get("key1")
	if !ok || val != "value1" {
		t.Error("key1 should still exist")
	}
}

func TestSetClearTTL(t *testing.T) {
	s, strStore := NewTestStringStore()
	ttlStore := ttl.New(s)

	strStore.Set("name", "Alice")
	ttlStore.Expire("name", 10)

	ttl, _ := ttlStore.TTL("name")
	if ttl <= 0 {
		t.Error("TTL should be set")
	}

	strStore.Set("name", "Bob")

	ttl, _ = ttlStore.TTL("name")
	if ttl != -1 {
		t.Error("TTL should be removed after SET")
	}
}

func TestGetBasic(t *testing.T) {
	_, strStore := NewTestStringStore()
	strStore.Set("name", "Alice")

	val, ok := strStore.Get("name")
	if !ok {
		t.Error("Key should exist")
	}
	if val != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", val)
	}
}

func TestGetNonExistent(t *testing.T) {
	_, strStore := NewTestStringStore()

	val, ok := strStore.Get("unknown")
	if ok {
		t.Error("Get should return ok=false for non-key exist")
	}
	if val != "" {
		t.Errorf("Expected empty string for wrong key, got '%s'", val)
	}
}

func TestGetWrongType(t *testing.T) {
	s, strStore := NewTestStringStore()

	key := "wrong_type_key"

	s.Mu.Lock()
	s.Data[key] = []string{"1", "2"}
	s.Types[key] = core.TypeList
	s.Mu.Unlock()

	val, ok := strStore.Get(key)
	if ok {
		t.Error("Get should return ok=false for non-string types")
	}
	if val != "" {
		t.Errorf("Expected empty string for wrong type, got '%s'", val)
	}
}

func TestGetClearsTTL(t *testing.T) {
	s, strStore := NewTestStringStore()
	ttlStore := ttl.New(s)

	strStore.Set("key", "value")
	ttlStore.Expire("key", 1)

	time.Sleep(1100 * time.Millisecond)

	_, ok := strStore.Get("key")
	if ok {
		t.Error("Get should return ok=false for expired key")
	}
}

func TestDeleteBasic(t *testing.T) {
	_, strStore := NewTestStringStore()
	strStore.Set("name", "Alice")

	strStore.Delete("name")

	val, ok := strStore.Get("name")
	if ok {
		t.Error("Key should be deleted, but is still exists")
	}
	if val != "" {
		t.Errorf("Expected empty string for deleted key, got '%s'", val)
	}
}

func TestDeleteNonExistent(t *testing.T) {
	_, strStore := NewTestStringStore()
	strStore.Delete("unknown")
}

func TestDeleteClearsTTL(t *testing.T) {
	s, strStore := NewTestStringStore()
	ttlStore := ttl.New(s)

	strStore.Set("key", "value")
	ttlStore.Expire("key", 10)

	strStore.Delete("key")

	ttl, _ := ttlStore.TTL("key")
	if ttl != -2 {
		t.Error("TTL must be cleared after Delete")
	}
}
