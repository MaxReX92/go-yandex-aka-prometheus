package storage

type StorageBackup interface {
	CreateBackup() error
	RestoreFromBackup() error
}
