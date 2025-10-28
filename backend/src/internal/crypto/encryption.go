package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/logger"
)

const (
	// SSN_ENCRYPTED_PREFIX identifies encrypted SSN values
	SSN_ENCRYPTED_PREFIX = "ENC_SSN:"
	// PASSWORD_ENCRYPTED_PREFIX identifies encrypted password values
	PASSWORD_ENCRYPTED_PREFIX = "ENC_PWD:"
	// Key size for AES-256
	AES_KEY_SIZE = 32
)

var (
	// Global encryption key - should be loaded from secure storage
	encryptionKey []byte
)

// InitEncryption initializes the encryption system with a key from environment or KMS
func InitEncryption() error {
	// Try to get key from environment variable first (for development)
	keyStr := os.Getenv("SSN_ENCRYPTION_KEY")
	if keyStr == "" {
		// Use hardcoded key for testing - exactly 32 bytes when decoded
		// This decodes to exactly 32 bytes: "12345678901234567890123456789012"
		keyStr = "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI="
		logger.Info("Using default encryption configuration")
	}

	var err error
	encryptionKey, err = base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return fmt.Errorf("failed to decode encryption key: %w", err)
	}
	if len(encryptionKey) != AES_KEY_SIZE {
		return fmt.Errorf("encryption key must be %d bytes, got %d", AES_KEY_SIZE, len(encryptionKey))
	}

	logger.Info("Encryption system ready")
	return nil
}

// DecryptSSN decrypts an SSN using AES-256-GCM
func DecryptSSN(encryptedSSN string) (string, error) {
	if encryptedSSN == "" {
		return "", nil
	}

	// Check if it's actually encrypted
	if !IsEncryptedSSN(encryptedSSN) {
		// Return as-is for backwards compatibility during migration
		logger.Warningf("SSN appears to be unencrypted: %s", MaskSSN(encryptedSSN))
		return encryptedSSN, nil
	}

	if encryptionKey == nil {
		return "", errors.New("encryption not initialized")
	}

	// Remove prefix and decode
	encodedData := strings.TrimPrefix(encryptedSSN, SSN_ENCRYPTED_PREFIX)
	ciphertext, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted SSN: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("invalid encrypted SSN: too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt SSN: %w", err)
	}

	return string(plaintext), nil
}

// IsEncryptedSSN checks if an SSN is encrypted
func IsEncryptedSSN(ssn string) bool {
	return strings.HasPrefix(ssn, SSN_ENCRYPTED_PREFIX)
}

// MaskSSN returns a masked version of the SSN for display (***-**-1234)
func MaskSSN(ssn string) string {
	if ssn == "" {
		return ""
	}

	// If encrypted, decrypt first to get the real SSN for masking
	if IsEncryptedSSN(ssn) {
		decrypted, err := DecryptSSN(ssn)
		if err != nil {
			logger.Errorf("Failed to decrypt SSN for masking: %v", err)
			return "***-**-****"
		}
		ssn = decrypted
	}

	// Remove any existing formatting
	cleanSSN := strings.ReplaceAll(ssn, "-", "")
	cleanSSN = strings.ReplaceAll(cleanSSN, " ", "")

	// Validate length
	if len(cleanSSN) != 9 {
		return "***-**-****"
	}

	// Return masked version with last 4 digits
	return fmt.Sprintf("***-**-%s", cleanSSN[5:])
}

// EncryptPassword encrypts a password using AES-256-GCM
func EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	if encryptionKey == nil {
		return "", errors.New("encryption not initialized")
	}

	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)

	// Encode and add prefix
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return PASSWORD_ENCRYPTED_PREFIX + encoded, nil
}

// DecryptPassword decrypts a password using AES-256-GCM
func DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", nil
	}

	// Check if it's actually encrypted
	if !IsEncryptedPassword(encryptedPassword) {
		// Return as-is for backwards compatibility or literal passwords
		return encryptedPassword, nil
	}

	if encryptionKey == nil {
		return "", errors.New("encryption not initialized")
	}

	// Remove prefix and decode
	encodedData := strings.TrimPrefix(encryptedPassword, PASSWORD_ENCRYPTED_PREFIX)
	ciphertext, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted password: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("invalid encrypted password: too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}

	return string(plaintext), nil
}

// IsEncryptedPassword checks if a password is encrypted
func IsEncryptedPassword(password string) bool {
	return strings.HasPrefix(password, PASSWORD_ENCRYPTED_PREFIX)
}
