package aeskit

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGCMNISTVectors(t *testing.T) {
	// NIST AES-GCM vectors with a zero AES-128 key and 96-bit nonce.
	for _, tc := range []struct {
		name  string
		plain []byte
		want  string
	}{
		{"empty", nil, "58e2fccefa7e3061367f1d57a4e7455a"},
		{"one_block", make([]byte, 16), "0388dace60b6a392f328c2b971b2fe78ab6e47d42cec13bdf53a67b21257bddf"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, nonce := make([]byte, 16), make([]byte, 12)
			want := decodeHex(t, tc.want)
			ct, err := EncryptGCM(key, nonce, tc.plain, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, want, ct.Bytes())
			assert.Equal(t, want[:len(want)-16], ct.Data())
			assert.Equal(t, want[len(want)-16:], ct.Tag())
			plain, err := DecryptGCM(key, nonce, want, nil, nil)
			require.NoError(t, err)
			assert.True(t, bytes.Equal(tc.plain, plain))
		})
	}
}

func TestGCMOptions(t *testing.T) {
	key, data, aad := make([]byte, 32), []byte("authenticated message"), []byte("metadata")
	block, err := aes.NewCipher(key)
	require.NoError(t, err)
	for _, tc := range []struct {
		name      string
		option    *GCMOption
		nonceSize int
		tagSize   int
	}{
		{"nil", nil, 12, 16},
		{"zero", &GCMOption{}, 12, 16},
		{"tag12", &GCMOption{TagSize: 12}, 12, 12},
		{"tag13", &GCMOption{TagSize: 13}, 12, 13},
		{"tag14", &GCMOption{TagSize: 14}, 12, 14},
		{"tag15", &GCMOption{TagSize: 15}, 12, 15},
		{"tag16", &GCMOption{TagSize: 16}, 12, 16},
		{"nonce1", &GCMOption{NonceSize: 1}, 1, 16},
		{"nonce8", &GCMOption{NonceSize: 8}, 8, 16},
		{"nonce16", &GCMOption{NonceSize: 16}, 16, 16},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nonce := bytes.Repeat([]byte{0x42}, tc.nonceSize)
			var aead cipher.AEAD
			if tc.tagSize != 16 {
				aead, err = cipher.NewGCMWithTagSize(block, tc.tagSize)
			} else {
				aead, err = cipher.NewGCMWithNonceSize(block, tc.nonceSize)
			}
			require.NoError(t, err)
			want := aead.Seal(nil, nonce, data, aad)
			ct, err := EncryptGCM(key, nonce, data, aad, tc.option)
			require.NoError(t, err)
			assert.Equal(t, want, ct.Bytes())
			assert.Len(t, ct.Data(), len(data))
			assert.Len(t, ct.Tag(), tc.tagSize)
			ciphertext := ct.Bytes()
			plain, err := DecryptGCM(key, nonce, ciphertext, aad, tc.option)
			require.NoError(t, err)
			assert.Equal(t, data, plain)
			assert.Equal(t, want, ciphertext)
			assert.Equal(t, []byte("authenticated message"), data)
			assert.Equal(t, []byte("metadata"), aad)
			assert.Equal(t, bytes.Repeat([]byte{0x42}, tc.nonceSize), nonce)
			assert.Equal(t, make([]byte, 32), key)
		})
	}
}

func TestGCMInvalidOptions(t *testing.T) {
	key := make([]byte, 16)
	for _, tc := range []struct {
		name      string
		option    *GCMOption
		nonceSize int
	}{
		{"negative_tag", &GCMOption{TagSize: -1}, 12},
		{"short_tag", &GCMOption{TagSize: 11}, 12},
		{"long_tag", &GCMOption{TagSize: 17}, 12},
		{"negative_nonce", &GCMOption{NonceSize: -1}, 12},
		{"custom_nonce_mismatch", &GCMOption{NonceSize: 8}, 12},
		{"custom_tag_nonce_mismatch", &GCMOption{TagSize: 12}, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nonce := make([]byte, tc.nonceSize)
			ct, err := EncryptGCM(key, nonce, nil, nil, tc.option)
			assert.Error(t, err)
			assert.Nil(t, ct)
			plain, err := DecryptGCM(key, nonce, nil, nil, tc.option)
			assert.Error(t, err)
			assert.Nil(t, plain)
		})
	}
}

func TestGCMRejectsUnauthenticatedData(t *testing.T) {
	for _, tagSize := range []int{12, 16} {
		t.Run(fmt.Sprintf("tag=%d", tagSize), func(t *testing.T) {
			key, nonce, aad := make([]byte, 16), make([]byte, 12), []byte("metadata")
			opt := &GCMOption{TagSize: tagSize}
			ct, err := EncryptGCM(key, nonce, []byte("secret"), aad, opt)
			require.NoError(t, err)
			changed := func(data []byte) []byte {
				out := bytes.Clone(data)
				out[0] ^= 1
				return out
			}
			badTag := ct.Bytes()
			badTag[len(badTag)-1] ^= 1
			for _, tc := range []struct {
				name                  string
				key, nonce, data, aad []byte
			}{
				{"key", changed(key), nonce, ct.Bytes(), aad},
				{"nonce", key, changed(nonce), ct.Bytes(), aad},
				{"aad", key, nonce, ct.Bytes(), changed(aad)},
				{"missing_aad", key, nonce, ct.Bytes(), nil},
				{"ciphertext", key, nonce, changed(ct.Bytes()), aad},
				{"tag", key, nonce, badTag, aad},
				{"truncated", key, nonce, ct.Bytes()[:len(ct.Bytes())-1], aad},
				{"shorter_than_tag", key, nonce, ct.Bytes()[:tagSize-1], aad},
				{"empty", key, nonce, nil, aad},
			} {
				t.Run(tc.name, func(t *testing.T) {
					original := bytes.Clone(tc.data)
					plain, err := DecryptGCM(tc.key, tc.nonce, tc.data, tc.aad, opt)
					assert.Error(t, err)
					assert.Nil(t, plain)
					assert.Equal(t, original, tc.data)
				})
			}
		})
	}
}
