package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// PKCS7Padding fills plaintext as an integral multiple of the block length
func PKCS7Padding(p []byte, blockSize int) []byte {
	pad := blockSize - len(p)%blockSize
	padtext := bytes.Repeat([]byte{byte(pad)}, pad)
	return append(p, padtext...)
}

// PKCS7UnPadding removes padding data from the tail of plaintext
func PKCS7UnPadding(p []byte) []byte {
	length := len(p)
	paddLen := int(p[length-1])
	return p[:(length - paddLen)]
}

// AESCBCEncrypt encrypts data with AES algorithm in CBC mode
// Note that key length must be 16, 24 or 32 bytes to select AES-128, AES-192, or AES-256
// Note that AES block size is 16 bytes
func AESCBCEncrypt(text, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	text = PKCS7Padding(text, block.BlockSize())
	ciphertext := make([]byte, len(text))
	blockMode := cipher.NewCBCEncrypter(block, key[:block.BlockSize()])
	blockMode.CryptBlocks(ciphertext, text)
	return ciphertext, nil
}

// AESCBCDecrypt decrypts cipher text with AES algorithm in CBC mode
// Note that key length must be 16, 24 or 32 bytes to select AES-128, AES-192, or AES-256
// Note that AES block size is 16 bytes
func AESCBCDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	blockMode := cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	blockMode.CryptBlocks(plaintext, ciphertext)
	return PKCS7UnPadding(plaintext), nil
}
