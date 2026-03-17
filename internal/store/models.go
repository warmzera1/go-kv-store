package store

type DataType int

const (
	TypeString DataType = iota
	TypeList
	TypeSet
	TypeHash
)

// String - метод для DataType
func (dt DataType) String() string {
	switch dt {
	case TypeString:
		return "string"
	case TypeList:
		return "list"
	case TypeSet:
		return "set"
	case TypeHash:
		return "hash"
	default:
		return "unknown"
	}
}

// ListValue - структура для хранения списка
type ListValue struct {
	Items []interface{}
}

// NewList - конструктор для списка
func NewList() *ListValue {
	return &ListValue{
		Items: make([]interface{}, 0),
	}
}

// SetValue - структура для хранения множества
type SetValue struct {
	Items map[interface{}]struct{}
}

// NewSet - конструктор для множества
func NewSet() *SetValue {
	return &SetValue{
		Items: make(map[interface{}]struct{}),
	}
}
