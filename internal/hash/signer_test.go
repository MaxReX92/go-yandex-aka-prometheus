package hash

import (
	"encoding/hex"
	"errors"
	"hash"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testSignerConfig struct {
	key string
}

type testHashHolder struct {
	hash         string
	errorMessage string
}

func TestSigner_GetSign(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		sighHash           string
		sighErrMessage     string
		expectedSign       []byte
		expectedErrMessage string
	}{
		{
			name:               "no_secret_key",
			expectedErrMessage: "failed to get signature: secret key was not initialized",
		},
		{
			name:               "signature_hash_error",
			key:                "test secret key",
			sighErrMessage:     "test get signature hash error message",
			expectedErrMessage: "failed to get signature hash: test get signature hash error message",
		},
		{
			name:         "success_signature",
			key:          "test secret key",
			sighHash:     "test signature hash",
			expectedSign: []byte("test signature hash"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := testSignerConfig{key: tt.key}
			holder := testHashHolder{hash: tt.sighHash, errorMessage: tt.sighErrMessage}
			signer := NewSigner(&conf)

			actualSign, actualErr := signer.GetSign(&holder)

			if tt.expectedSign != nil {
				assert.Equal(t, tt.expectedSign, actualSign)
			}

			if tt.expectedErrMessage == "" {
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, tt.expectedErrMessage)
			}
		})
	}
}

func TestSigner_GetSignString(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		sighHash           string
		sighErrMessage     string
		expectedSign       string
		expectedErrMessage string
	}{
		{
			name:               "no_secret_key",
			expectedErrMessage: "failed to get signature: secret key was not initialized",
		},
		{
			name:               "signature_hash_error",
			key:                "test secret key",
			sighErrMessage:     "test get signature hash error message",
			expectedErrMessage: "failed to get signature hash: test get signature hash error message",
		},
		{
			name:         "success_signature",
			key:          "test secret key",
			sighHash:     "test signature hash",
			expectedSign: "test signature hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := testSignerConfig{key: tt.key}
			holder := testHashHolder{hash: tt.sighHash, errorMessage: tt.sighErrMessage}
			signer := NewSigner(&conf)

			actualSign, actualErr := signer.GetSignString(&holder)

			if tt.expectedSign != "" {
				actual, err := hex.DecodeString(actualSign)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSign, string(actual))
			}

			if tt.expectedErrMessage == "" {
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, tt.expectedErrMessage)
			}
		})
	}
}

func TestSigner_CheckSign(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		sighHash           string
		sighErrMessage     string
		sign               []byte
		expectedOk         bool
		expectedErrMessage string
	}{
		{
			name:               "no_secret_key",
			expectedErrMessage: "failed to get signature: secret key was not initialized",
		},
		{
			name:               "signature_hash_error",
			key:                "test secret key",
			sighErrMessage:     "test get signature hash error message",
			expectedErrMessage: "failed to get signature hash: test get signature hash error message",
		},
		{
			name:     "other_signature",
			key:      "test secret key",
			sighHash: "test signature hash",
			sign:     []byte("other test signature hash"),
		},
		{
			name:       "same_signature",
			key:        "test secret key",
			sighHash:   "test signature hash",
			sign:       []byte("test signature hash"),
			expectedOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := testSignerConfig{key: tt.key}
			holder := testHashHolder{hash: tt.sighHash, errorMessage: tt.sighErrMessage}
			signer := NewSigner(&conf)

			actualOk, actualErr := signer.CheckSign(&holder, tt.sign)

			assert.Equal(t, tt.expectedOk, actualOk)

			if tt.expectedErrMessage == "" {
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, tt.expectedErrMessage)
			}
		})
	}
}

func TestSigner_CheckSignString(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		sighHash           string
		sighErrMessage     string
		sign               string
		expectedOk         bool
		expectedErrMessage string
	}{
		{
			name:               "no_secret_key",
			expectedErrMessage: "failed to get signature: secret key was not initialized",
		},
		{
			name:               "signature_hash_error",
			key:                "test secret key",
			sighErrMessage:     "test get signature hash error message",
			expectedErrMessage: "failed to get signature hash: test get signature hash error message",
		},
		{
			name:               "invalid_signature",
			key:                "test secret key",
			sighHash:           "test signature hash",
			sign:               "invalid test signature hash",
			expectedErrMessage: "failed to decode signature: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name:     "other_signature",
			key:      "test secret key",
			sighHash: "test signature hash",
			sign:     hex.EncodeToString([]byte("other test signature hash")),
		},
		{
			name:       "same_signature",
			key:        "test secret key",
			sighHash:   "test signature hash",
			sign:       hex.EncodeToString([]byte("test signature hash")),
			expectedOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := testSignerConfig{key: tt.key}
			holder := testHashHolder{hash: tt.sighHash, errorMessage: tt.sighErrMessage}
			signer := NewSigner(&conf)

			actualOk, actualErr := signer.CheckSignString(&holder, tt.sign)

			assert.Equal(t, tt.expectedOk, actualOk)

			if tt.expectedErrMessage == "" {
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, tt.expectedErrMessage)
			}
		})
	}
}

func (t *testSignerConfig) GetKey() []byte {
	if t.key == "" {
		return nil
	}
	return []byte(t.key)
}

func (t *testHashHolder) GetHash(hash hash.Hash) ([]byte, error) {
	var err error
	if t.errorMessage != "" {
		err = errors.New(t.errorMessage) //nolint:goerr113
	}

	return []byte(t.hash), err
}
