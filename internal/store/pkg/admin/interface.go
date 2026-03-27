package admin

type AdminInterface interface {
	SaveSnapshot(filename string) error
	LoadSnapshot(filename string) error
	BGSave(filename string) error
	FlushDB() error
}
