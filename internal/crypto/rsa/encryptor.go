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

type rsaEncryptor struct {
	publicKey *rsa.PublicKey
}

func NewEncryptor(publicCertPath string) (*rsaEncryptor, error) {
	publicCert, err := os.ReadFile(publicCertPath)
	if err != nil {
		return nil, logger.WrapError("read public cert file", err)
	}

	publicPem, _ := pem.Decode(publicCert)
	if publicPem == nil || publicPem.Type != "PUBLIC KEY" {
		return nil, logger.WrapError("decode PEM block containing public key", crypto.ErrInvalidKey)
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicPem.Bytes)
	if err != nil {
		return nil, logger.WrapError("parse public key", err)
	}

	return &rsaEncryptor{
		publicKey: publicKey.(*rsa.PublicKey),
	}, nil
}

func (r *rsaEncryptor) Encrypt(bytes []byte) ([]byte, error) {
	var result []byte
	hash := sha256.New()
	blockSize := r.publicKey.Size()
	for _, messageChunk := range chunk.SliceToChunks(bytes, blockSize) {
		encryptedBlock, err := rsa.EncryptOAEP(hash, rand.Reader, r.publicKey, messageChunk, nil)
		if err != nil {
			return nil, logger.WrapError("encrypt message", err)
		}
		result = append(result, encryptedBlock...)
	}

	return result, nil
}
