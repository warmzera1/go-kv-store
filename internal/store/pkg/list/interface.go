package list

type ListInterface interface {
	LPush(key string, values ...interface{}) error
	RPush(key string, values ...interface{}) error
	RPop(key string) (interface{}, error)
	LPop(key string) (interface{}, error)
	LLen(key string) (int, error)
	LIndex(key string, index int) (interface{}, error)
	LRange(key string, start, stop int) ([]interface{}, error)
}
