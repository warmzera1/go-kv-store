package set

type SetInterface interface {
	SAdd(key string, values ...interface{}) (int, error)
	SRem(key string, values ...interface{}) (int, error)
	SIsMember(key string, value interface{}) (bool, error)
	SMembers(key string) ([]interface{}, error)
	SCard(key string) (int, error)
}
