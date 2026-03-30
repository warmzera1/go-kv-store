package test

import (
	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/list"
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
