package storage

type StorageBackup interface {
	Create()
	Restore()
}
