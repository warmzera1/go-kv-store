package hash

type HashInterface interface {
	HSet(key string, field string, value interface{}) error
	HGet(key string, field string) (interface{}, error)
	HGetAll(key string) (map[string]interface{}, error)
	HDel(key string, fields ...string) error
}
