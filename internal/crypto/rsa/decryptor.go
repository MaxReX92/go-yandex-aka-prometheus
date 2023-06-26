package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/pkg/chunk"
)

type rsaDecryptor struct {
	privateKey *rsa.PrivateKey
}

func NewDecryptor(privateKeyPAth string) (*rsaDecryptor, error) {
	privateKeyContent, err := os.ReadFile(privateKeyPAth)
	if err != nil {
		return nil, logger.WrapError("read private key file", err)
	}

	privatePem, _ := pem.Decode(privateKeyContent)
	if privatePem == nil || privatePem.Type != "RSA PRIVATE KEY" {
		return nil, logger.WrapError("decode PEM block containing private key", crypto.ErrInvalidKey)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privatePem.Bytes)
	if err != nil {
		return nil, logger.WrapError("parse public key", err)
	}

	return &rsaDecryptor{
		privateKey: privateKey,
	}, nil
}

func (r *rsaDecryptor) Decrypt(bytes []byte) ([]byte, error) {
	var result []byte
	hash := sha256.New()
	blockSize := r.privateKey.Size()
	for _, messageChunk := range chunk.Chunk(bytes, blockSize) {
		decryptedBlock, err := rsa.DecryptOAEP(hash, rand.Reader, r.privateKey, messageChunk, nil)
		if err != nil {
			return nil, logger.WrapError("decrypt message", err)
		}
		result = append(result, decryptedBlock...)
	}

	return result, nil
}
