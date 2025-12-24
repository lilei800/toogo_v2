// Package encrypt
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description AES-256加密工具
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"hotgo/internal/consts"
)

var (
	ErrInvalidKeyLength   = errors.New("invalid key length, must be 32 bytes for AES-256")
	ErrInvalidCiphertext  = errors.New("invalid ciphertext")
	ErrDecryptionFailed   = errors.New("decryption failed")
)

// AESEncryptor AES-256加密器
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor 创建AES加密器
func NewAESEncryptor(key string) (*AESEncryptor, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != consts.EncryptionKeyLength {
		// 如果密钥长度不足，进行填充或截断
		keyBytes = padOrTruncateKey(keyBytes, consts.EncryptionKeyLength)
	}
	return &AESEncryptor{key: keyBytes}, nil
}

// padOrTruncateKey 填充或截断密钥到指定长度
func padOrTruncateKey(key []byte, length int) []byte {
	if len(key) >= length {
		return key[:length]
	}
	padded := make([]byte, length)
	copy(padded, key)
	return padded
}

// Encrypt 加密字符串
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// 使用GCM模式（更安全）
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Base64编码
	return consts.ApiKeyPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密字符串
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// 检查是否有加密前缀
	if !strings.HasPrefix(ciphertext, consts.ApiKeyPrefix) {
		// 未加密的字符串，直接返回
		return ciphertext, nil
	}

	// 移除前缀
	ciphertext = strings.TrimPrefix(ciphertext, consts.ApiKeyPrefix)

	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// IsEncrypted 检查字符串是否已加密
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, consts.ApiKeyPrefix)
}

// 全局加密器实例
var defaultEncryptor *AESEncryptor

// InitEncryptor 初始化全局加密器
func InitEncryptor(key string) error {
	enc, err := NewAESEncryptor(key)
	if err != nil {
		return err
	}
	defaultEncryptor = enc
	return nil
}

// EncryptApiKey 加密API密钥
func EncryptApiKey(apiKey string) (string, error) {
	if defaultEncryptor == nil {
		if err := InitEncryptor(consts.DefaultEncryptionKey); err != nil {
			return "", err
		}
	}
	return defaultEncryptor.Encrypt(apiKey)
}

// DecryptApiKey 解密API密钥
func DecryptApiKey(encryptedKey string) (string, error) {
	if defaultEncryptor == nil {
		if err := InitEncryptor(consts.DefaultEncryptionKey); err != nil {
			return "", err
		}
	}
	return defaultEncryptor.Decrypt(encryptedKey)
}

// MaskApiKey 遮蔽API密钥（用于显示）
func MaskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// MaskSecretKey 遮蔽Secret密钥
func MaskSecretKey(secretKey string) string {
	if len(secretKey) <= 8 {
		return "********"
	}
	return secretKey[:4] + "********" + secretKey[len(secretKey)-4:]
}

// AesDecrypt 解密字符串（兼容旧接口）
func AesDecrypt(ciphertext string) (string, error) {
	return DecryptApiKey(ciphertext)
}

// AesEncrypt 加密字符串（兼容旧接口）
func AesEncrypt(plaintext string) (string, error) {
	return EncryptApiKey(plaintext)
}

// AesECBDecrypt AES ECB模式解密（用于密码等）
func AesECBDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	blockSize := block.BlockSize()
	if len(ciphertext) < blockSize {
		return nil, errors.New("ciphertext too short")
	}
	if len(ciphertext)%blockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}
	
	plaintext := make([]byte, len(ciphertext))
	for bs, be := 0, blockSize; bs < len(ciphertext); bs, be = bs+blockSize, be+blockSize {
		block.Decrypt(plaintext[bs:be], ciphertext[bs:be])
	}
	
	// 去除PKCS7填充
	plaintext = pkcs7UnPadding(plaintext)
	return plaintext, nil
}

// AesECBEncrypt AES ECB模式加密
func AesECBEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	blockSize := block.BlockSize()
	plaintext = pkcs7Padding(plaintext, blockSize)
	
	ciphertext := make([]byte, len(plaintext))
	for bs, be := 0, blockSize; bs < len(plaintext); bs, be = bs+blockSize, be+blockSize {
		block.Encrypt(ciphertext[bs:be], plaintext[bs:be])
	}
	
	return ciphertext, nil
}

// pkcs7Padding PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7UnPadding PKCS7去除填充
func pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}
	unpadding := int(data[length-1])
	if unpadding > length {
		return data
	}
	return data[:(length - unpadding)]
}
