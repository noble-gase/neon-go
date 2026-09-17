package aeskit

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type aesMode struct {
	name    string
	ivSize  int
	encrypt func(key, iv, data []byte) (*CipherText, error)
	decrypt func(key, iv, data []byte) ([]byte, error)
}

var aesModes = []aesMode{
	{"CBC", 16,
		func(key, iv, data []byte) (*CipherText, error) { return EncryptCBC(key, iv, data) },
		func(key, iv, data []byte) ([]byte, error) { return DecryptCBC(key, iv, data) }},
	{"ECB", 0,
		func(key, iv, data []byte) (*CipherText, error) { return EncryptECB(key, data) },
		func(key, iv, data []byte) ([]byte, error) { return DecryptECB(key, data) }},
	{"CTR", 16, EncryptCTR, DecryptCTR},
	{"GCM", 12,
		func(key, iv, data []byte) (*CipherText, error) { return EncryptGCM(key, iv, data, nil, nil) },
		func(key, iv, data []byte) ([]byte, error) { return DecryptGCM(key, iv, data, nil, nil) }},
}

func decodeHex(t *testing.T, value string) []byte {
	t.Helper()
	data, err := hex.DecodeString(value)
	require.NoError(t, err)
	return data
}

func TestAESNISTVectors(t *testing.T) {
	// NIST SP 800-38A, first two blocks of the AES-128 examples.
	key := decodeHex(t, "2b7e151628aed2a6abf7158809cf4f3c")
	plain := decodeHex(t, "6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51")
	for _, tc := range []struct {
		mode aesMode
		iv   string
		want string
	}{
		{aesModes[0], "000102030405060708090a0b0c0d0e0f", "7649abac8119b246cee98e9b12e9197d5086cb9b507219ee95db113a917678b2"},
		{aesModes[1], "", "3ad77bb40d7a3660a89ecaf32466ef97f5d3d58503b9699de785895a96fdbaaf"},
		{aesModes[2], "f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff", "874d6191b620e3261bef6864990db6ce9806f66b7970fdff8617187bb9fffdff"},
	} {
		t.Run(tc.mode.name, func(t *testing.T) {
			iv, want := decodeHex(t, tc.iv), decodeHex(t, tc.want)
			ct, err := tc.mode.encrypt(key, iv, plain)
			require.NoError(t, err)
			// CBC and ECB append a full PKCS#7 block to the standard vector.
			require.GreaterOrEqual(t, len(ct.Bytes()), len(want))
			assert.Equal(t, want, ct.Bytes()[:len(want)])
			got, err := tc.mode.decrypt(key, iv, ct.Bytes())
			require.NoError(t, err)
			assert.Equal(t, plain, got)
		})
	}
}

func TestAESRoundTripBoundaries(t *testing.T) {
	for _, mode := range aesModes {
		for _, keySize := range []int{16, 24, 32} {
			for _, size := range []int{0, 1, 15, 16, 17, 31, 32, 257} {
				t.Run(fmt.Sprintf("%s/key=%d/length=%d", mode.name, keySize, size), func(t *testing.T) {
					key := bytes.Repeat([]byte{0x42}, keySize)
					iv := bytes.Repeat([]byte{0x24}, mode.ivSize)
					// Spare capacity makes accidental append-to-input writes observable.
					backing := bytes.Repeat([]byte{0xa5}, size+32)
					data := backing[:size]
					ct, err := mode.encrypt(key, iv, data)
					require.NoError(t, err)
					wantLen := size
					switch mode.name {
					case "CBC", "ECB":
						wantLen += aes.BlockSize - size%aes.BlockSize
					case "GCM":
						wantLen += 16
					}
					assert.Len(t, ct.Bytes(), wantLen)
					ciphertext := ct.Bytes()
					original := bytes.Clone(ciphertext)
					got, err := mode.decrypt(key, iv, ciphertext)
					require.NoError(t, err)
					assert.True(t, bytes.Equal(data, got))
					assert.Equal(t, original, ciphertext)
					assert.Equal(t, bytes.Repeat([]byte{0xa5}, size+32), backing)
					assert.Equal(t, bytes.Repeat([]byte{0x42}, keySize), key)
					assert.Equal(t, bytes.Repeat([]byte{0x24}, mode.ivSize), iv)
					if mode.name != "GCM" {
						assert.Equal(t, ct.Bytes(), ct.Data())
						assert.Empty(t, ct.Tag())
					}
				})
			}
		}
	}
}

func TestAESInvalidKeysAndIVs(t *testing.T) {
	for _, mode := range aesModes {
		t.Run(mode.name, func(t *testing.T) {
			for _, size := range []int{0, 15, 17, 23, 25, 31, 33} {
				t.Run(fmt.Sprintf("key=%d", size), func(t *testing.T) {
					key, iv := make([]byte, size), make([]byte, mode.ivSize)
					ct, err := mode.encrypt(key, iv, nil)
					assert.ErrorIs(t, err, aes.KeySizeError(size))
					assert.Nil(t, ct)
					plain, err := mode.decrypt(key, iv, nil)
					assert.ErrorIs(t, err, aes.KeySizeError(size))
					assert.Nil(t, plain)
				})
			}
			if mode.ivSize == 0 {
				return
			}
			for _, size := range []int{0, mode.ivSize - 1, mode.ivSize + 1} {
				t.Run(fmt.Sprintf("iv=%d", size), func(t *testing.T) {
					key, iv := make([]byte, 16), make([]byte, size)
					ct, err := mode.encrypt(key, iv, nil)
					assert.Error(t, err)
					assert.Nil(t, ct)
					plain, err := mode.decrypt(key, iv, nil)
					assert.Error(t, err)
					assert.Nil(t, plain)
				})
			}
		})
	}
}

func TestAESRejectsInvalidPadding(t *testing.T) {
	key, iv := make([]byte, 16), make([]byte, 16)
	block, err := aes.NewCipher(key)
	require.NoError(t, err)
	for _, mode := range aesModes[:2] {
		t.Run(mode.name, func(t *testing.T) {
			for _, size := range []int{0, 1, 15, 17, 31} {
				plain, err := mode.decrypt(key, iv, make([]byte, size))
				assert.Error(t, err)
				assert.Nil(t, plain)
			}
			for _, tail := range [][]byte{{0}, {17}, {1, 2}} {
				// Encrypt intentionally invalid padding without using our padding helper.
				padded := make([]byte, 16)
				copy(padded[len(padded)-len(tail):], tail)
				data := make([]byte, 16)
				if mode.name == "CBC" {
					cipher.NewCBCEncrypter(block, iv).CryptBlocks(data, padded)
				} else {
					block.Encrypt(data, padded)
				}
				plain, err := mode.decrypt(key, iv, data)
				assert.EqualError(t, err, "aeskit: invalid padding")
				assert.Nil(t, plain)
			}
		})
	}
}
