package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func (s *Signer) GetSignString(holder HashHolder) (string, error) {
	sign, err := s.GetSign(holder)
	if err != nil {
		return "", logger.WrapError("get sign", err)
	}

	return hex.EncodeToString(sign), nil
}

func (s *Signer) GetSign(holder HashHolder) ([]byte, error) {
	if s.hash == nil {
		return nil, logger.WrapError("get signature", ErrMissedSecretKey)
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	defer s.hash.Reset()

	return holder.GetHash(s.hash)
}

func (s *Signer) CheckSign(holder HashHolder, signature string) (bool, error) {
	sign, err := hex.DecodeString(signature)
	if err != nil {
		return false, logger.WrapError("decode signature", err)
	}

	holderSign, err := s.GetSign(holder)
	if err != nil {
		return false, logger.WrapError("get holder hash", err)
	}

	return hmac.Equal(holderSign, sign), nil
}
