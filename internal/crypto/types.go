package crypto

type Encryptor interface {
	Encrypt(bytes []byte) ([]byte, error)
}

type Decryptor interface {
	Decrypt(bytes []byte) ([]byte, error)
}
