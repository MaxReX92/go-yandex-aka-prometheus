package hash

import "hash"

type HashHolder interface {
	GetHash(hash hash.Hash) ([]byte, error)
}
