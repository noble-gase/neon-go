package deskit

import (
	"crypto/des"
	"errors"
)

func pkcs7_padding(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 || blockSize > 255 || blockSize%des.BlockSize != 0 {
		return nil, errors.New("deskit: invalid padding size")
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
	if blockSize <= 0 || blockSize > 255 || blockSize%des.BlockSize != 0 {
		return nil, errors.New("deskit: invalid padding size")
	}

	length := len(data)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("deskit: invalid data length")
	}

	pad := int(data[length-1])
	if pad == 0 || pad > blockSize {
		return nil, errors.New("deskit: invalid padding")
	}

	for _, b := range data[length-pad:] {
		if b != byte(pad) {
			return nil, errors.New("deskit: invalid padding")
		}
	}
	return data[:length-pad], nil
}
