package test

import (
	"testing"
	"time"

	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

func TestExpireBasic(t *testing.T) {
	s, strStore := NewTestStringStore()
	ttlStore := ttl.New(s)

	strStore.Set("key", "value")

	ttlStore.Expire("key", 1)

	ttl, err := ttlStore.TTL("key")
	if err != nil {
		t.Fatalf("Failed TTL: %v", err)
	}
	if ttl <= 0 {
		t.Errorf("Expeceted TTL > 0, got %d", ttl)
	}

	time.Sleep(1100 * time.Millisecond)

	val, ok := strStore.Get("key")
	if ok {
		t.Error("Key should be expired")
	}
	if val != "" {
		t.Errorf("Expected empty string, got '%s'", val)
	}

	ttl, err = ttlStore.TTL("key")
	if err != nil {
		t.Fatalf("TTL Failed: %v", err)
	}
	if ttl != -2 {
		t.Errorf("Expected TTL -2 (key not exist), got %d", ttl)
	}
}

func TestTTL(t *testing.T) {
	s, strStore := NewTestStringStore()
	ttlStore := ttl.New(s)

	strStore.Set("key", "value")

	ttlStore.Expire("key", 5)

	ttl, err := ttlStore.TTL("key")
	if err != nil {
		t.Fatalf("TTL Failed: %v", err)
	}
	if ttl <= 0 || ttl > 5 {
		t.Errorf("Expected TTL ~ 5, got %d", ttl)
	}

	time.Sleep(2 * time.Second)

	ttl, err = ttlStore.TTL("key")
	if err != nil {
		t.Fatalf("TTL Failed: %v", err)
	}
	if ttl <= 0 || ttl > 4 {
		t.Errorf("Expected TTL ~ 3, got %d", ttl)
	}

	_, ok := strStore.Get("key")
	if !ok {
		t.Error("Key should still exist")
	}

	time.Sleep(3 * time.Second)

	_, ok = strStore.Get("key")
	if ok {
		t.Error("Key should be expired")
	}

	ttl, err = ttlStore.TTL("key")
	if err != nil {
		t.Fatalf("TTL Failed: %v", err)
	}
	if ttl != -2 {
		t.Errorf("Expected TTL -2, got %d", ttl)
	}
}

func TestPersist(t *testing.T) {
	s, strStore := NewTestStringStore()
	ttlStore := ttl.New(s)

	strStore.Set("key", "value")

	ttlStore.Expire("key", 10)

	ttl, err := ttlStore.TTL("key")
	if err != nil {
		t.Fatalf("TTL Failed: %v", err)
	}
	if ttl < 0 {
		t.Errorf("Expected TTL > 0, got %d", ttl)
	}

	ok, err := ttlStore.Persist("key")
	if err != nil {
		t.Fatalf("Failed Persist: %v", err)
	}
	if !ok {
		t.Error("Expected removed TTL")
	}

	ttl, err = ttlStore.TTL("key")
	if err != nil {
		t.Fatalf("TTL Failed: %v", err)
	}
	if ttl != -1 {
		t.Errorf("Expected TTL -1, (ttl not exist), got %d", ttl)
	}

	time.Sleep(2 * time.Second)

	_, ok = strStore.Get("key")
	if !ok {
		t.Error("Key should stiil exist after Persist")
	}

	ttl, err = ttlStore.TTL("key")
	if err != nil {
		t.Fatalf("TTL Failed: %v", err)
	}
	if ttl != -1 {
		t.Errorf("Expected TTL -1, (ttl not exist), got %d", ttl)
	}
}

func TestExpireNonExistent(t *testing.T) {
	s, strStore := NewTestStringStore()
	ttlStore := ttl.New(s)

	ok, err := ttlStore.Expire("noneexist", 1)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}
	if ok {
		t.Error("Expire should return false for non-existent key")
	}

	_, exists := strStore.Get("nonexist")
	if exists {
		t.Error("Key should not exist")
	}

	ttl, err := ttlStore.TTL("noneexist")
	if err != nil {
		t.Fatalf("TTL Failed: %v", err)
	}
	if ttl != -2 {
		t.Errorf("Expected TTL -2, got %d", ttl)
	}
}
