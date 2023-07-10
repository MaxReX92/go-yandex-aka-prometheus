package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
)

// SignerConfig contains required Signer settings.
type SignerConfig interface {
	GetKey() []byte
}

// Signer provide sign functional.
type Signer struct {
	hash hash.Hash
	lock sync.Mutex
}

// NewSigner create new instance on Signer.
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

// GetSignString returns signed string.
func (s *Signer) GetSignString(holder HashHolder) (string, error) {
	sign, err := s.GetSign(holder)
	if err != nil {
		return "", logger.WrapError("get sign", err)
	}

	return hex.EncodeToString(sign), nil
}

// GetSign returns object signature.
func (s *Signer) GetSign(holder HashHolder) ([]byte, error) {
	if s.hash == nil {
		return nil, logger.WrapError("get signature", ErrMissedSecretKey)
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	defer s.hash.Reset()

	return holder.GetHash(s.hash)
}

// CheckSignString validate object signature string.
func (s *Signer) CheckSignString(holder HashHolder, signature string) (bool, error) {
	sign, err := hex.DecodeString(signature)
	if err != nil {
		return false, logger.WrapError("decode signature", err)
	}

	return s.CheckSign(holder, sign)
}

// CheckSign validate object signature.
func (s *Signer) CheckSign(holder HashHolder, signature []byte) (bool, error) {
	holderSign, err := s.GetSign(holder)
	if err != nil {
		return false, logger.WrapError("get holder hash", err)
	}

	return hmac.Equal(holderSign, signature), nil
}
