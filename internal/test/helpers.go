package test

import (
	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/hash"
	"github.com/warmzera1/kv-store/internal/store/pkg/list"
	"github.com/warmzera1/kv-store/internal/store/pkg/set"
	str "github.com/warmzera1/kv-store/internal/store/pkg/string"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

func NewTestStringStore() (*core.Store, *str.StringStore) {
	s := core.New()
	ttlStore := ttl.New(s)
	strStore := str.New(s, ttlStore)

	return s, strStore
}

func NewTestListStore() (*core.Store, *list.ListStore) {
	s := core.New()
	ttlStore := ttl.New(s)
	listStore := list.New(s, ttlStore)

	return s, listStore
}

func NewTestSetStore() (*core.Store, *set.SetStore) {
	s := core.New()
	ttlStore := ttl.New(s)
	setStore := set.New(s, ttlStore)

	return s, setStore
}

func NewTestHashStore() (*core.Store, *hash.HashStore) {
	s := core.New()
	ttlStore := ttl.New(s)
	hashStore := hash.New(s, ttlStore)

	return s, hashStore
}
