package test

import (
	"os"
	"testing"

	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/admin"
	"github.com/warmzera1/kv-store/internal/store/pkg/hash"
	"github.com/warmzera1/kv-store/internal/store/pkg/list"
	"github.com/warmzera1/kv-store/internal/store/pkg/set"
	str "github.com/warmzera1/kv-store/internal/store/pkg/string"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

func TestFlushDB(t *testing.T) {
	s := core.New()
	adminStore := admin.New(s)
	ttlStore := ttl.New(s)

	strStore := str.New(s, ttlStore)
	setStore := set.New(s, ttlStore)
	listStore := list.New(s, ttlStore)
	hashStore := hash.New(s, ttlStore)

	strStore.Set("str-key", "value")
	setStore.SAdd("set-key", "value")
	listStore.LPush("list-key", "value")
	hashStore.HSet("hash-key", "field", "value")

	err := adminStore.FlushDB()
	if err != nil {
		t.Fatalf("Failed FlushDB: %v", err)
	}

	_, ok := strStore.Get("str-key")
	if ok {
		t.Fatalf("String key should be deleted")
	}

	count, _ := setStore.SCard("set-key")
	if count != 0 {
		t.Error("List key shoud be deleted")
	}

	length, _ := listStore.LLen("list-key")
	if length != 0 {
		t.Error("List key should be deleted")
	}

	fields, _ := hashStore.HGetAll("hash-key")
	if len(fields) != 0 {
		t.Error("Hash key should be deleted")
	}
}
func TestExistsBasic(t *testing.T) {
	s := core.New()
	adminStore := admin.New(s)
	ttlStore := ttl.New(s)
	strStore := str.New(s, ttlStore)

	strStore.Set("key", "value")

	count := adminStore.Exists("key")
	if count != 1 {
		t.Fatalf("Expeceted 1, got %d", count)
	}

	count = adminStore.Exists("noneexist")
	if count != 0 {
		t.Errorf("Expected 0, got %d", count)
	}
}

func TestExistsMultiple(t *testing.T) {
	s := core.New()
	adminStore := admin.New(s)
	ttlStore := ttl.New(s)
	strStore := str.New(s, ttlStore)

	strStore.Set("key1", "value1")
	strStore.Set("key2", "value2")

	count := adminStore.Exists("key1", "key2")
	if count != 2 {
		t.Fatalf("Expected 2, got %d", count)
	}

	count = adminStore.Exists("key3", "key4")
	if count != 0 {
		t.Fatalf("Expected 0, got %d", count)
	}
}

func TestExistsNonExist(t *testing.T) {
	s := core.New()
	adminStore := admin.New(s)

	err := adminStore.Exists("key3", "key4")
	if err != 0 {
		t.Fatalf("Failed Exists: %v", err)
	}
}

func TestSaveLoadSnapshot(t *testing.T) {
	s := core.New()
	ttlStore := ttl.New(s)

	strStore := str.New(s, ttlStore)
	setStore := set.New(s, ttlStore)
	listStore := list.New(s, ttlStore)
	hashStore := hash.New(s, ttlStore)

	strStore.Set("str-key", "value")
	setStore.SAdd("set-key", "value")
	listStore.LPush("list-key", "value")
	hashStore.HSet("hash-key", "field", "value")

	err := s.SaveSnapshot("test-snapshot.gob")
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	newS := core.New()
	err = newS.LoadSnapshot("test-snapshot.gob")
	if err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}

	newTtlStore := ttl.New(newS)
	newStrStore := str.New(newS, newTtlStore)
	newSetStore := set.New(newS, newTtlStore)
	newListStore := list.New(newS, newTtlStore)
	newHashStore := hash.New(newS, newTtlStore)

	val, ok := newStrStore.Get("str-key")
	if !ok {
		t.Error("String key not restored")
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %s", val)
	}

	count, _ := newSetStore.SCard("set-key")
	if count != 1 {
		t.Error("Set key not restored")
	}

	length, _ := newListStore.LLen("list-key")
	if length != 1 {
		t.Error("List key not restored")
	}

	fields, _ := newHashStore.HGetAll("hash-key")
	if len(fields) != 1 {
		t.Error("Hash key not restored")
	}

	os.Remove("test-snapshot.gob")
}
