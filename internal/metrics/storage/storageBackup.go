package storage

import (
	"context"
	"io"
)

type StorageBackup interface {
	io.Closer

	CreateBackup(context.Context) error
	RestoreFromBackup(context.Context) error
}
