// Package deskit 封装 DES 的 CBC、ECB、CTR、CFB、OFB 加解密操作。
// 密钥长度为 8 字节，CBC、CTR、CFB 和 OFB 的 IV 长度为 8 字节。
package deskit

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"errors"
)

// CipherText 加密文本
type CipherText struct {
	bytes []byte
}

// String 返回加密数据base64字符串
func (ct *CipherText) String() string {
	return base64.StdEncoding.EncodeToString(ct.bytes)
}

// Bytes 返回加密数据的副本
func (ct *CipherText) Bytes() []byte {
	return bytes.Clone(ct.bytes)
}

// ------------------------------------ DES-CBC ------------------------------------

// EncryptCBC DES-CBC 加密(pkcs#7, 默认填充BlockSize；paddingSize须为BlockSize的倍数且不超过255)
func EncryptCBC(key, iv, data []byte, paddingSize ...int) (*CipherText, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	blockSize := block.BlockSize()
	if len(paddingSize) != 0 {
		blockSize = paddingSize[0]
	}
	paddingBytes, err := pkcs7_padding(data, blockSize)
	if err != nil {
		return nil, err
	}

	bm := cipher.NewCBCEncrypter(block, iv)
	if len(paddingBytes)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(paddingBytes))
	bm.CryptBlocks(out, paddingBytes)

	return &CipherText{bytes: out}, nil
}

// DecryptCBC DES-CBC 解密(pkcs#7, 默认填充BlockSize；paddingSize须为BlockSize的倍数且不超过255；若加密时使用了自定义 paddingSize，解密需传入相同值)
func DecryptCBC(key, iv, data []byte, paddingSize ...int) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	bm := cipher.NewCBCDecrypter(block, iv)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	blockSize := block.BlockSize()
	if len(paddingSize) != 0 {
		blockSize = paddingSize[0]
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return pkcs7_unpadding(out, blockSize)
}

// ------------------------------------ DES-ECB ------------------------------------

// EncryptECB DES-ECB 加密(pkcs#7, 默认填充BlockSize；paddingSize须为BlockSize的倍数且不超过255)
func EncryptECB(key, data []byte, paddingSize ...int) (*CipherText, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(paddingSize) != 0 {
		blockSize = paddingSize[0]
	}
	paddingBytes, err := pkcs7_padding(data, blockSize)
	if err != nil {
		return nil, err
	}

	bm := NewECBEncrypter(block)
	if len(paddingBytes)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(paddingBytes))
	bm.CryptBlocks(out, paddingBytes)

	return &CipherText{bytes: out}, nil
}

// DecryptECB DES-ECB 解密(pkcs#7, 默认填充BlockSize；paddingSize须为BlockSize的倍数且不超过255；若加密时使用了自定义 paddingSize，解密需传入相同值)
func DecryptECB(key, data []byte, paddingSize ...int) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bm := NewECBDecrypter(block)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	blockSize := block.BlockSize()
	if len(paddingSize) != 0 {
		blockSize = paddingSize[0]
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return pkcs7_unpadding(out, blockSize)
}

// ------------------------------------ DES-CTR ------------------------------------

// EncryptCTR DES-CTR 加密
func EncryptCTR(key, iv, data []byte) (*CipherText, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(out, data)

	return &CipherText{bytes: out}, nil
}

// DecryptCTR DES-CTR 解密
func DecryptCTR(key, iv, data []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}

// ------------------------------------ DES-CFB ------------------------------------

// EncryptCFB DES-CFB 加密(CFB64，无填充)
func EncryptCFB(key, iv, data []byte) (*CipherText, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(out, data)

	return &CipherText{bytes: out}, nil
}

// DecryptCFB DES-CFB 解密(CFB64，无填充)
func DecryptCFB(key, iv, data []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}

// ------------------------------------ DES-OFB ------------------------------------

// EncryptOFB DES-OFB 加密(无填充)
func EncryptOFB(key, iv, data []byte) (*CipherText, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))
	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(out, data)

	return &CipherText{bytes: out}, nil
}

// DecryptOFB DES-OFB 解密(无填充)
func DecryptOFB(key, iv, data []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))
	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}
