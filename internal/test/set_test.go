package test

import (
	"testing"

	"github.com/warmzera1/kv-store/internal/store/core"
)

func TestSAddSingle(t *testing.T) {
	_, setStore := NewTestSetStore()

	added, err := setStore.SAdd("key", "a")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 1 {
		t.Errorf("Expected item 1, got %d", added)
	}

	length, err := setStore.SCard("key")
	if err != nil {
		t.Fatalf("SCard failed: %v", length)
	}
	if length != 1 {
		t.Errorf("Expected length 1, got %d", length)
	}

	val, err := setStore.SIsMember("key", "a")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if !val {
		t.Errorf("Expected val is true")
	}
}

func TestSAddMultiple(t *testing.T) {
	_, setStore := NewTestSetStore()

	item, err := setStore.SAdd("key", "a", "b", "c")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if item != 3 {
		t.Errorf("Expected item 3, got %d", item)
	}

	length, err := setStore.SCard("key")
	if err != nil {
		t.Fatalf("SCard failed: %v", length)
	}
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}

	expected := []string{"a", "b", "c"}
	for _, v := range expected {
		ok, err := setStore.SIsMember("key", v)
		if err != nil {
			t.Fatalf("SIsMember failed: %v", err)
		}
		if !ok {
			t.Errorf("Expected number '%s' to exist", v)
		}
	}
}

func TestSAddDuplicate(t *testing.T) {
	_, setStore := NewTestSetStore()

	added, err := setStore.SAdd("key", "a", "b")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 2 {
		t.Errorf("Expected item 2, got %d", added)
	}

	added, err = setStore.SAdd("key", "a", "c")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 1 {
		t.Errorf("Expected item 1, got %d", added)
	}

	length, err := setStore.SCard("key")
	if err != nil {
		t.Fatalf("SCard failed: %v", length)
	}
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}

	val, err := setStore.SIsMember("key", "a")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if !val {
		t.Errorf("Expected val is true")
	}

	val, err = setStore.SIsMember("key", "b")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if !val {
		t.Errorf("Expected val is true")
	}

	val, err = setStore.SIsMember("key", "c")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if !val {
		t.Errorf("Expected val is true")
	}
}

func TestSRemBasic(t *testing.T) {
	_, setStore := NewTestSetStore()

	setStore.SAdd("key", "a", "b", "c")
	item, err := setStore.SRem("key", "b")
	if err != nil {
		t.Fatalf("SRem failed: %v", err)
	}
	if item != 1 {
		t.Errorf("Expected removed item 1, got %d", item)
	}

	length, err := setStore.SCard("key")
	if err != nil {
		t.Fatalf("SCard failed: %v", length)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	val, err := setStore.SIsMember("key", "b")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if val {
		t.Errorf("Expected val is false")
	}

	expected := []string{"a", "c"}
	for _, v := range expected {
		ok, err := setStore.SIsMember("key", v)
		if err != nil {
			t.Fatalf("Failed SIsMember: %v", err)
		}
		if !ok {
			t.Errorf("Expected number '%s' of exist", v)
		}
	}
}

func TestSRemNonExistent(t *testing.T) {
	_, setStore := NewTestSetStore()

	setStore.SAdd("key", "a", "b")
	item, err := setStore.SRem("key", "c")
	if err != nil {
		t.Fatalf("SRem failed: %v", err)
	}
	if item != 0 {
		t.Errorf("Expected removed item 0, got %d", item)
	}

	length, err := setStore.SCard("key")
	if err != nil {
		t.Fatalf("SCard failed: %v", length)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}
}

func TestSIsMember(t *testing.T) {
	_, setStore := NewTestSetStore()

	added, err := setStore.SAdd("key", "a", "b")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 2 {
		t.Errorf("Expected item 2, got %d", added)
	}

	val, err := setStore.SIsMember("key", "a")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if !val {
		t.Errorf("Expected val is true")
	}

	val, err = setStore.SIsMember("key", "b")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if !val {
		t.Errorf("Expected val is true")
	}

	val, err = setStore.SIsMember("noneexist", "x")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if val {
		t.Errorf("Expected val is false")
	}
}

func TestSMembers(t *testing.T) {
	_, setStore := NewTestSetStore()

	setStore.SAdd("key", "a", "b")
	members, err := setStore.SMembers("key")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}

	expected := map[string]bool{"a": true, "b": true}
	for _, m := range members {
		str, ok := m.(string)
		if !ok {
			t.Errorf("Members is not string: %v", m)
		}
		if !expected[str] {
			t.Errorf("Unexpected member: %s", str)
		}
	}

	members, err = setStore.SMembers("nonexist")
	if err != nil {
		t.Fatalf("SMembers failed: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("Expected empty slice, got %d", len(members))
	}
}

func TestSCard(t *testing.T) {
	_, setStore := NewTestSetStore()

	setStore.SAdd("key", "a", "b")

	length, err := setStore.SCard("key")
	if err != nil {
		t.Fatalf("SCard failed: %v", length)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	item, err := setStore.SRem("key", "b")
	if err != nil {
		t.Fatalf("SRem failed: %v", err)
	}
	if item != 1 {
		t.Errorf("Expected removed item 1, got %d", item)
	}

	length, err = setStore.SCard("key")
	if err != nil {
		t.Fatalf("SCard failed: %v", length)
	}
	if length != 1 {
		t.Errorf("Expected length 1, got %d", length)
	}

	length, err = setStore.SCard("noneexist")
	if err != nil {
		t.Fatalf("SCard failed: %v", length)
	}
	if length != 0 {
		t.Errorf("Expected length 0, got %d", length)
	}
}

func TestSAddWrongType(t *testing.T) {
	s, setStore := NewTestSetStore()

	key := "wrong_type_key"

	s.Mu.Lock()
	s.Data[key] = "set"
	s.Types[key] = core.TypeString
	s.Mu.Unlock()

	_, err := setStore.SAdd(key, "a", "b")
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}
