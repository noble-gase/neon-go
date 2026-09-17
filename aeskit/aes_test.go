package aeskit

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESCompatibilityVectors(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv, nonce := key[:16], key[:12]
	data, aad := []byte("ILoveNobleGase"), []byte("IIInsomnia")

	// Preserve the original fixed ciphertexts, including custom padding and GCM AAD.
	// GCM options and authentication failures are covered in gcm_test.go.
	for _, tc := range []struct {
		name    string
		encrypt func() (*CipherText, error)
		decrypt func([]byte) ([]byte, error)
		want    string
		tagSize int
	}{
		{
			name:    "CBC/default_padding",
			encrypt: func() (*CipherText, error) { return EncryptCBC(key, iv, data) },
			decrypt: func(ciphertext []byte) ([]byte, error) { return DecryptCBC(key, iv, ciphertext) },
			want:    "WDq8s1qdHCML8YLhfdmGRw==",
		},
		{
			name:    "CBC/padding_32",
			encrypt: func() (*CipherText, error) { return EncryptCBC(key, iv, data, 32) },
			decrypt: func(ciphertext []byte) ([]byte, error) { return DecryptCBC(key, iv, ciphertext, 32) },
			want:    "vjemH/hxbwNh+WXhkKseCu2GrM4O6bnaaKv59wgkRSE=",
		},
		{
			name:    "ECB/default_padding",
			encrypt: func() (*CipherText, error) { return EncryptECB(key, data) },
			decrypt: func(ciphertext []byte) ([]byte, error) { return DecryptECB(key, ciphertext) },
			want:    "oYDjdGHY8lK1/sJo750Waw==",
		},
		{
			name:    "ECB/padding_32",
			encrypt: func() (*CipherText, error) { return EncryptECB(key, data, 32) },
			decrypt: func(ciphertext []byte) ([]byte, error) { return DecryptECB(key, ciphertext, 32) },
			want:    "u0iDWHM8JMnRyJNCiCzKJNib2cOjUrx2FqMjmg3ZTZA=",
		},
		{
			name:    "CTR",
			encrypt: func() (*CipherText, error) { return EncryptCTR(key, iv, data) },
			decrypt: func(ciphertext []byte) ([]byte, error) { return DecryptCTR(key, iv, ciphertext) },
			want:    "KP7OnZj9J9ONnjn6yA0=",
		},
		{
			name:    "GCM/with_aad",
			encrypt: func() (*CipherText, error) { return EncryptGCM(key, nonce, data, aad, &GCMOption{}) },
			decrypt: func(ciphertext []byte) ([]byte, error) { return DecryptGCM(key, nonce, ciphertext, aad, nil) },
			want:    "qciumnROL4U9F0klEKhzE/DngAy/clYUsZGfcafh",
			tagSize: 16,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, err := base64.StdEncoding.DecodeString(tc.want)
			require.NoError(t, err)

			t.Run("encrypt", func(t *testing.T) {
				ct, err := tc.encrypt()
				require.NoError(t, err)
				require.NotNil(t, ct)
				assert.Equal(t, tc.want, ct.String())
				assert.Equal(t, want, ct.Bytes())
				assert.Equal(t, want[:len(want)-tc.tagSize], ct.Data())
				assert.Equal(t, want[len(want)-tc.tagSize:], ct.Tag())
			})

			t.Run("decrypt", func(t *testing.T) {
				// Decrypt the fixed fixture independently of the encryption result.
				plain, err := tc.decrypt(want)
				require.NoError(t, err)
				assert.Equal(t, data, plain)
			})
		})
	}
}

type aesMode struct {
	name    string
	ivSize  int
	encrypt func(key, iv, data []byte, paddingSize ...int) (*CipherText, error)
	decrypt func(key, iv, data []byte, paddingSize ...int) ([]byte, error)
}

var aesModes = []aesMode{
	{"CBC", aes.BlockSize, EncryptCBC, DecryptCBC},
	{"ECB", 0,
		func(key, iv, data []byte, paddingSize ...int) (*CipherText, error) {
			return EncryptECB(key, data, paddingSize...)
		},
		func(key, iv, data []byte, paddingSize ...int) ([]byte, error) {
			return DecryptECB(key, data, paddingSize...)
		}},
	{"CTR", aes.BlockSize,
		func(key, iv, data []byte, paddingSize ...int) (*CipherText, error) { return EncryptCTR(key, iv, data) },
		func(key, iv, data []byte, paddingSize ...int) ([]byte, error) { return DecryptCTR(key, iv, data) }},
	{"GCM", 12,
		func(key, iv, data []byte, paddingSize ...int) (*CipherText, error) {
			return EncryptGCM(key, iv, data, nil, nil)
		},
		func(key, iv, data []byte, paddingSize ...int) ([]byte, error) {
			return DecryptGCM(key, iv, data, nil, nil)
		}},
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
		paddings := []int{0} // 0 means omit the optional paddingSize argument.
		if mode.name == "CBC" || mode.name == "ECB" {
			paddings = append(paddings, 16, 32, 240)
		}
		for _, keySize := range []int{16, 24, 32} {
			for _, padding := range paddings {
				for _, size := range []int{0, 1, 15, 16, 17, 31, 32, 33, 239, 240, 241, 257} {
					t.Run(fmt.Sprintf("%s/key=%d/padding=%d/length=%d", mode.name, keySize, padding, size), func(t *testing.T) {
						key := bytes.Repeat([]byte{0x42}, keySize)
						iv := bytes.Repeat([]byte{0x24}, mode.ivSize)
						var paddingSize []int
						if padding != 0 {
							paddingSize = []int{padding}
						}
						// Spare capacity makes accidental append-to-input writes observable.
						backing := bytes.Repeat([]byte{0xa5}, size+256)
						data := backing[:size]
						ct, err := mode.encrypt(key, iv, data, paddingSize...)
						require.NoError(t, err)
						wantLen := size
						switch mode.name {
						case "CBC", "ECB":
							blockSize := padding
							if blockSize == 0 {
								blockSize = aes.BlockSize
							}
							wantLen += blockSize - size%blockSize
						case "GCM":
							wantLen += 16
						}
						ciphertext := ct.Bytes()
						assert.Len(t, ciphertext, wantLen)
						assert.Equal(t, base64.StdEncoding.EncodeToString(ciphertext), ct.String())
						original := bytes.Clone(ciphertext)
						got, err := mode.decrypt(key, iv, ciphertext, paddingSize...)
						require.NoError(t, err)
						assert.True(t, bytes.Equal(data, got))
						assert.Equal(t, original, ciphertext)
						assert.Equal(t, bytes.Repeat([]byte{0xa5}, size+256), backing)
						assert.Equal(t, bytes.Repeat([]byte{0x42}, keySize), key)
						assert.Equal(t, bytes.Repeat([]byte{0x24}, mode.ivSize), iv)
						if mode.name != "GCM" {
							assert.Equal(t, ciphertext, ct.Data())
							assert.Empty(t, ct.Tag())
						}
					})
				}
			}
		}
	}
}

func TestAESMismatchedPaddingSize(t *testing.T) {
	key, iv := make([]byte, 16), make([]byte, aes.BlockSize)
	for _, mode := range aesModes[:2] {
		t.Run(mode.name, func(t *testing.T) {
			t.Run("ciphertext_not_multiple_of_padding_size", func(t *testing.T) {
				ct, err := mode.encrypt(key, iv, []byte("hello"))
				require.NoError(t, err)
				plain, err := mode.decrypt(key, iv, ct.Bytes(), 32)
				assert.EqualError(t, err, "aeskit: invalid data length")
				assert.Nil(t, plain)
			})
			t.Run("padding_exceeds_configured_size", func(t *testing.T) {
				ct, err := mode.encrypt(key, iv, nil, 32)
				require.NoError(t, err)
				plain, err := mode.decrypt(key, iv, ct.Bytes())
				assert.EqualError(t, err, "aeskit: invalid padding")
				assert.Nil(t, plain)
			})
		})
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

func TestCipherTextReturnsCopies(t *testing.T) {
	for _, tc := range []struct {
		name string
		get  func(*CipherText) []byte
		want []byte
	}{
		{"Bytes", (*CipherText).Bytes, []byte{1, 2, 3, 4}},
		{"Data", (*CipherText).Data, []byte{1, 2}},
		{"Tag", (*CipherText).Tag, []byte{3, 4}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ct := &CipherText{bytes: []byte{1, 2, 3, 4}, tagsize: 2}
			original := ct.String()
			got := tc.get(ct)
			assert.Equal(t, tc.want, got)
			for i := range got {
				got[i] ^= 0xff
			}
			assert.Equal(t, tc.want, tc.get(ct))
			assert.Equal(t, []byte{1, 2, 3, 4}, ct.Bytes())
			assert.Equal(t, original, ct.String())
		})
	}
}
