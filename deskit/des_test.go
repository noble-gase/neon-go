package deskit

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCipherTextBytesReturnsCopy(t *testing.T) {
	ct := &CipherText{bytes: []byte{1, 2, 3, 4}}
	original := ct.String()
	got := ct.Bytes()
	assert.Equal(t, []byte{1, 2, 3, 4}, got)
	for i := range got {
		got[i] ^= 0xff
	}
	assert.Equal(t, []byte{1, 2, 3, 4}, ct.Bytes())
	assert.Equal(t, original, ct.String())
}

type desMode struct {
	name    string
	encrypt func(key, iv, data []byte, paddingSize ...int) (*CipherText, error)
	decrypt func(key, iv, data []byte, paddingSize ...int) ([]byte, error)
}

var desModes = []desMode{
	{"CBC", EncryptCBC, DecryptCBC},
	{"ECB",
		func(key, iv, data []byte, paddingSize ...int) (*CipherText, error) {
			return EncryptECB(key, data, paddingSize...)
		},
		func(key, iv, data []byte, paddingSize ...int) ([]byte, error) {
			return DecryptECB(key, data, paddingSize...)
		}},
	{"CTR",
		func(key, iv, data []byte, paddingSize ...int) (*CipherText, error) {
			return EncryptCTR(key, iv, data)
		},
		func(key, iv, data []byte, paddingSize ...int) ([]byte, error) {
			return DecryptCTR(key, iv, data)
		}},
	{"CFB",
		func(key, iv, data []byte, paddingSize ...int) (*CipherText, error) {
			return EncryptCFB(key, iv, data)
		},
		func(key, iv, data []byte, paddingSize ...int) ([]byte, error) {
			return DecryptCFB(key, iv, data)
		}},
	{"OFB",
		func(key, iv, data []byte, paddingSize ...int) (*CipherText, error) {
			return EncryptOFB(key, iv, data)
		},
		func(key, iv, data []byte, paddingSize ...int) ([]byte, error) {
			return DecryptOFB(key, iv, data)
		}},
}

func decodeHex(t *testing.T, value string) []byte {
	t.Helper()
	data, err := hex.DecodeString(value)
	require.NoError(t, err)
	return data
}

func TestDESKnownVectors(t *testing.T) {
	// Standard DES vector: E(133457799BBCDFF1, 0123456789ABCDEF).
	key := decodeHex(t, "133457799bbcdff1")
	plain := decodeHex(t, "0123456789abcdef")
	want := decodeHex(t, "85e813540f0ab405")
	ct, err := EncryptECB(key, plain)
	require.NoError(t, err)
	assert.Equal(t, want, ct.Bytes()[:des.BlockSize])

	// CTR encrypts the initial counter to produce its first keystream block.
	ct, err = EncryptCTR(key, plain, make([]byte, des.BlockSize))
	require.NoError(t, err)
	assert.Equal(t, want, ct.Bytes())
	got, err := DecryptCTR(key, plain, want)
	require.NoError(t, err)
	assert.Equal(t, make([]byte, des.BlockSize), got)

	// FIPS 81 CBC example; PKCS#7 adds another block after these 24 bytes.
	key = decodeHex(t, "0123456789abcdef")
	iv := decodeHex(t, "1234567890abcdef")
	plain = []byte("Now is the time for all ")
	want = decodeHex(t, "e5c7cdde872bf27c43e934008c389c0f683788499a7c05f6")
	ct, err = EncryptCBC(key, iv, plain)
	require.NoError(t, err)
	assert.Equal(t, want, ct.Bytes()[:len(plain)])
}

func TestDESFeedbackVectors(t *testing.T) {
	// Verified with LibreSSL enc -des-cfb / -des-ofb using -nopad.
	// The message spans multiple blocks and ends with a partial block.
	key := decodeHex(t, "0123456789abcdef")
	iv := decodeHex(t, "1234567890abcdef")
	plain := []byte("Now is the time for all")
	for _, tc := range []struct {
		name    string
		encrypt func(key, iv, data []byte) (*CipherText, error)
		decrypt func(key, iv, data []byte) ([]byte, error)
		want    string
	}{
		{"CFB", EncryptCFB, DecryptCFB, "f3096249c7f46e51a69e839b1a92f78403467133898ea6"},
		{"OFB", EncryptOFB, DecryptOFB, "f3096249c7f46e5135f24a242eeb3d3f3d6d5be3255af8"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := decodeHex(t, tc.want)
			ct, err := tc.encrypt(key, iv, plain)
			require.NoError(t, err)
			assert.Equal(t, want, ct.Bytes())
			got, err := tc.decrypt(key, iv, want)
			require.NoError(t, err)
			assert.Equal(t, plain, got)
		})
	}
}

func TestDESRoundTrip(t *testing.T) {
	for _, mode := range desModes {
		for _, size := range []int{0, 1, 7, 8, 9, 16, 31, 32, 257} {
			for _, padding := range []int{8, 16, 32} {
				t.Run(fmt.Sprintf("%s/length=%d/padding=%d", mode.name, size, padding), func(t *testing.T) {
					key, iv := []byte("12345678"), []byte("abcdefgh")
					data := bytes.Repeat([]byte{0xa5}, size)
					original := bytes.Clone(data)
					var paddingSize []int
					if padding != des.BlockSize {
						paddingSize = []int{padding}
					}
					ct, err := mode.encrypt(key, iv, data, paddingSize...)
					require.NoError(t, err)
					wantLen := size
					if mode.name == "CBC" || mode.name == "ECB" {
						wantLen += padding - size%padding
					}
					assert.Len(t, ct.Bytes(), wantLen)
					assert.Equal(t, base64.StdEncoding.EncodeToString(ct.Bytes()), ct.String())
					cipherOriginal := bytes.Clone(ct.Bytes())
					got, err := mode.decrypt(key, iv, ct.Bytes(), paddingSize...)
					require.NoError(t, err)
					assert.Equal(t, data, got)
					assert.Equal(t, original, data)
					assert.Equal(t, cipherOriginal, ct.Bytes())
					assert.Equal(t, []byte("12345678"), key)
					assert.Equal(t, []byte("abcdefgh"), iv)
				})
			}
		}
	}
}

func TestDESInvalidInputs(t *testing.T) {
	key, iv := []byte("12345678"), []byte("abcdefgh")
	for _, mode := range desModes {
		t.Run(mode.name, func(t *testing.T) {
			for _, size := range []int{0, 7, 9, 16, 24} {
				_, err := mode.encrypt(make([]byte, size), iv, nil)
				assert.Error(t, err)
				_, err = mode.decrypt(make([]byte, size), iv, nil)
				assert.Error(t, err)
			}
			if mode.name != "ECB" {
				for _, size := range []int{0, 7, 9, 16} {
					_, err := mode.encrypt(key, make([]byte, size), nil)
					assert.Error(t, err)
					_, err = mode.decrypt(key, make([]byte, size), nil)
					assert.Error(t, err)
				}
			}
			if mode.name != "CBC" && mode.name != "ECB" {
				return
			}
			for _, size := range []int{0, 1, 7, 9} {
				_, err := mode.decrypt(key, iv, make([]byte, size))
				assert.Error(t, err)
			}
			for _, padding := range []int{-1, 0, 256} {
				_, err := mode.encrypt(key, iv, nil, padding)
				assert.Error(t, err)
				_, err = mode.decrypt(key, iv, make([]byte, des.BlockSize), padding)
				assert.Error(t, err)
			}
			_, err := mode.encrypt(key, iv, nil, 7)
			assert.Error(t, err)

			ct, err := mode.encrypt(key, iv, []byte("hello"))
			require.NoError(t, err)
			_, err = mode.decrypt(key, iv, ct.Bytes(), 16)
			assert.Error(t, err)
		})
	}
}

func TestDESRejectsInvalidPadding(t *testing.T) {
	key, iv := []byte("12345678"), []byte("abcdefgh")
	block, err := des.NewCipher(key)
	require.NoError(t, err)
	for _, plain := range [][]byte{
		{1, 2, 3, 4, 5, 6, 7, 0}, // zero padding length
		{1, 2, 3, 4, 5, 6, 7, 9}, // padding exceeds block size
		{1, 2, 3, 4, 5, 6, 1, 2}, // inconsistent padding bytes
	} {
		for _, mode := range desModes[:2] {
			t.Run(fmt.Sprintf("%s/%x", mode.name, plain), func(t *testing.T) {
				// Encrypt raw invalid padding to exercise public decryption APIs.
				data := make([]byte, des.BlockSize)
				if mode.name == "CBC" {
					cipher.NewCBCEncrypter(block, iv).CryptBlocks(data, plain)
				} else {
					block.Encrypt(data, plain)
				}
				got, err := mode.decrypt(key, iv, data)
				assert.Error(t, err)
				assert.Nil(t, got)
			})
		}
	}
}
