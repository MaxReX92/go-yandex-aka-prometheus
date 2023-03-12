package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"hash"
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
)

type SignerConfig interface {
	GetKey() []byte
}

type Signer struct {
	hash hash.Hash
	lock sync.Mutex
}

func NewSigner(config SignerConfig) *Signer {
	var h hash.Hash
	key := config.GetKey()
	if key != nil {
		h = hmac.New(sha256.New, key)
	}
	return &Signer{
		hash: h,
	}
}

func (s *Signer) GetSign(holder HashHolder) ([]byte, error) {
	if s.hash == nil {
		return nil, errors.New("secret key was not initialized")
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	defer s.hash.Reset()

	return holder.GetHash(s.hash)
}

func (s *Signer) CheckSign(holder HashHolder, sign []byte) (bool, error) {
	holderSign, err := s.GetSign(holder)
	if err != nil {
		logger.ErrorFormat("Fail to get holder hash: %v", err)
		return false, err
	}

	return hmac.Equal(holderSign, sign), nil
}
