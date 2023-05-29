package hash

import "hash"

// HashHolder implementations serves the hash function.
type HashHolder interface {
	// GetHash returns a hash code for the current object.
	GetHash(hash hash.Hash) ([]byte, error)
}
