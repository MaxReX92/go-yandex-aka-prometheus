package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
)

type SignerConfig interface {
	GetKey() []byte
}

type Signer struct {
	hash hash.Hash
	sync.Mutex
}

func NewSigner(config SignerConfig) *Signer {
	return &Signer{
		hash: hmac.New(sha256.New, config.GetKey()),
	}
}

func (s *Signer) GetSign(holder HashHolder) ([]byte, error) {
	s.Lock()
	defer s.Unlock()
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
