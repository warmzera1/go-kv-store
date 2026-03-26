package admin

type AdminInterface interface {
	SaveSnapshot(filename string) error
	LoadSnapshot(filename string) error
}
