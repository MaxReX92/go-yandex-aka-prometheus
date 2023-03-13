package storage

import "io"

type StorageBackup interface {
	io.Closer

	CreateBackup() error
	RestoreFromBackup() error
}
