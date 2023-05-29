package storage

import (
	"context"
	"io"
)

// StorageBackup provide functionality for create or restore backups.
type StorageBackup interface {
	io.Closer

	// CreateBackup creates metrics backup.
	CreateBackup(context.Context) error

	// RestoreFromBackup recover state from backup.
	RestoreFromBackup(context.Context) error
}
