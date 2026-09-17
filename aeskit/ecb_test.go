package aeskit

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestECBBlockMode(t *testing.T) {
	// NIST SP 800-38A AES-128 ECB vector, without padding.
	key := decodeHex(t, "2b7e151628aed2a6abf7158809cf4f3c")
	plain := decodeHex(t, "6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51")
	encrypted := decodeHex(t, "3ad77bb40d7a3660a89ecaf32466ef97f5d3d58503b9699de785895a96fdbaaf")
	block, err := aes.NewCipher(key)
	require.NoError(t, err)
	for _, tc := range []struct {
		name string
		mode cipher.BlockMode
		src  []byte
		want []byte
	}{
		{"encrypt", NewECBEncrypter(block), plain, encrypted},
		{"decrypt", NewECBDecrypter(block), encrypted, plain},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, aes.BlockSize, tc.mode.BlockSize())
			src := bytes.Clone(tc.src)
			dst := bytes.Repeat([]byte{0xa5}, len(src)+aes.BlockSize)
			tc.mode.CryptBlocks(dst, src)
			assert.Equal(t, tc.want, dst[:len(src)])
			assert.Equal(t, bytes.Repeat([]byte{0xa5}, aes.BlockSize), dst[len(src):])
			assert.Equal(t, tc.src, src)
			tc.mode.CryptBlocks(src, src)
			assert.Equal(t, tc.want, src)
			assert.NotPanics(t, func() { tc.mode.CryptBlocks(nil, nil) })
			assert.PanicsWithValue(t, "crypto/cipher: input not full blocks", func() {
				tc.mode.CryptBlocks(make([]byte, 16), make([]byte, 15))
			})
			assert.PanicsWithValue(t, "crypto/cipher: output smaller than input", func() {
				tc.mode.CryptBlocks(make([]byte, 15), make([]byte, 16))
			})
		})
	}
}
