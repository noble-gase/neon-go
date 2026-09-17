package aeskit

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaddingSizeValidation(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, aes.BlockSize)
	for _, mode := range []struct {
		name    string
		encrypt func([]byte, ...int) (*CipherText, error)
		decrypt func([]byte, ...int) ([]byte, error)
	}{
		{"CBC",
			func(data []byte, padding ...int) (*CipherText, error) { return EncryptCBC(key, iv, data, padding...) },
			func(data []byte, padding ...int) ([]byte, error) { return DecryptCBC(key, iv, data, padding...) }},
		{"ECB",
			func(data []byte, padding ...int) (*CipherText, error) { return EncryptECB(key, data, padding...) },
			func(data []byte, padding ...int) ([]byte, error) { return DecryptECB(key, data, padding...) }},
	} {
		t.Run(mode.name, func(t *testing.T) {
			// A one-byte pad yields three complete cipher blocks. Previously,
			// padding sizes 1 and 1.5*BlockSize could succeed with this input.
			plain := bytes.Repeat([]byte{0x42}, 3*aes.BlockSize-1)
			ct, err := mode.encrypt(plain)
			require.NoError(t, err)
			for _, padding := range []int{-aes.BlockSize, 0, 1, aes.BlockSize - 1, aes.BlockSize * 3 / 2, 255, 256} {
				t.Run(fmt.Sprintf("invalid=%d", padding), func(t *testing.T) {
					encrypted, err := mode.encrypt(plain, padding)
					assert.EqualError(t, err, "aeskit: invalid padding size")
					assert.Nil(t, encrypted)
					decrypted, err := mode.decrypt(ct.Bytes(), padding)
					assert.EqualError(t, err, "aeskit: invalid padding size")
					assert.Nil(t, decrypted)
				})
			}
			for _, padding := range []int{aes.BlockSize, 2 * aes.BlockSize, 255 / aes.BlockSize * aes.BlockSize} {
				t.Run(fmt.Sprintf("valid=%d", padding), func(t *testing.T) {
					for _, data := range [][]byte{nil, plain} {
						encrypted, err := mode.encrypt(data, padding)
						require.NoError(t, err)
						decrypted, err := mode.decrypt(encrypted.Bytes(), padding)
						require.NoError(t, err)
						assert.True(t, bytes.Equal(data, decrypted))
					}
				})
			}
		})
	}
}
