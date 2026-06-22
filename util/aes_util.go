package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"
)

var (
	AESCbcEncryptionKey []byte
)

var (
	encryptionKey = []byte("8TFzOsqymO6Wc3e6") // 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
)

func AesCbcEncrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(AESCbcEncryptionKey)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	padLength := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padLength)}, padLength)
	data = append(data, pad...)

	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(data))
	mode.CryptBlocks(ciphertext, data)

	encrypted := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func AesCbcDecrypt(encrypted string) ([]byte, error) {
	block, err := aes.NewCipher(AESCbcEncryptionKey)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	iv := decoded[:blockSize]
	ciphertext := decoded[blockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	padLength := int(plaintext[len(plaintext)-1])
	plaintext = plaintext[:len(plaintext)-padLength]

	return plaintext, nil
}

// func MaskEmail(email string) string {
// 	parts := strings.Split(email, "@")
// 	if len(parts) != 2 {
// 		return email // Return original if invalid
// 	}
// 	username := parts[0]
// 	domain := parts[1]
// 	// Show at least 3 characters of the username
// 	visibleChars := 3
// 	if len(username) < 3 {
// 		visibleChars = len(username) // Show full username if it's too short
// 	}
// 	maskedUsername := username[:visibleChars] + strings.Repeat("*", len(username)-visibleChars)
// 	// Mask domain but keep TLD visible
// 	maskedDomain := "*****" + domain[strings.Index(domain, "."):]
// 	return maskedUsername + "@" + maskedDomain
// }

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email // Return original if invalid
	}
	username := parts[0]
	domain := parts[1]

	// Show at least 3 characters of the username
	visibleChars := 3
	if len(username) < 3 {
		visibleChars = len(username)
	}

	maskedUsername := username[:visibleChars] + strings.Repeat("*", len(username)-visibleChars)

	return maskedUsername + "@" + domain
}

func MaskMobile(mobile string) string {
	if len(mobile) < 4 {
		return "****"
	}
	return "******" + mobile[len(mobile)-5:]
}

var emailIV = []byte("0123456789abcdef") // Fixed IV (16 bytes)
func EncryptWithFixedIv(email string) string {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return ""
	}
	// Pad the email to match the block size
	blockSize := block.BlockSize()
	padLength := blockSize - len(email)%blockSize
	pad := bytes.Repeat([]byte{byte(padLength)}, padLength)
	emailBytes := append([]byte(email), pad...)
	// Encrypt using CBC mode with a fixed IV
	mode := cipher.NewCBCEncrypter(block, emailIV)
	ciphertext := make([]byte, len(emailBytes))
	mode.CryptBlocks(ciphertext, emailBytes)
	// Encode as base64 and return
	return base64.StdEncoding.EncodeToString(ciphertext)
}
func DecryptWithFixedIv(encryptedEmail string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedEmail)
	if err != nil {
		return "", err
	}
	// Decrypt using the fixed IV
	mode := cipher.NewCBCDecrypter(block, emailIV)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	// Remove padding
	padLength := int(plaintext[len(plaintext)-1])
	plaintext = plaintext[:len(plaintext)-padLength]
	return string(plaintext), nil
}

func HashString(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))

	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
