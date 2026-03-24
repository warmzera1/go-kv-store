package str

// InterfaceString - определяет методы для работы со строками
type InterfaceString interface {
	Set(key, value string)
	Get(key string) (string, bool)
	Delete(key string)
}
