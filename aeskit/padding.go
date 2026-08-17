package aeskit

import (
	"bytes"
	"errors"
)

func pkcs7_padding(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, errors.New("aeskit: invalid block size")
	}

	pad := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+pad)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out, nil
}

func pkcs7_unpadding(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, errors.New("aeskit: invalid block size")
	}

	length := len(data)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("aeskit: invalid data length")
	}

	pad := int(data[length-1])
	if pad == 0 || pad > blockSize {
		return nil, errors.New("aeskit: invalid padding")
	}
	if !bytes.Equal(data[length-pad:], bytes.Repeat([]byte{byte(pad)}, pad)) {
		return nil, errors.New("aeskit: invalid padding")
	}
	return data[:length-pad], nil
}
