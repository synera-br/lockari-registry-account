package cryptclient

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"lockari-api-application/pkg/utils"
)

// Crypt interface defines the methods for cryptographic operations
type Crypt interface {
	EncryptData(data interface{}) ([]byte, error)
	DecryptData(data []byte) ([]byte, error)
	ValidateKey(key []byte) bool
	GenerateKey() ([]byte, error)
}

// CryptData holds the configuration for cryptographic operations
type CryptData struct {
	EncryptionKey []byte
	token         *string
}

// NewCryptData creates a new CryptData instance with proper key generation
func NewCryptData(token *string) (*CryptData, error) {
	if token == nil {
		return nil, errors.New("crypt token cannot be nil")
	}

	if len(*token) == 0 {
		return nil, errors.New("crypt token cannot be empty")
	}

	c := &CryptData{
		token: token,
	}

	// Generate encryption key from token
	key, err := c.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	c.EncryptionKey = key
	return c, nil
}

// GenerateKey generates a 32-byte encryption key from the token using SHA-256
func (cd *CryptData) GenerateKey() ([]byte, error) {
	if cd.token == nil || len(*cd.token) == 0 {
		return nil, errors.New("invalid token for key generation")
	}

	// Use SHA-256 to generate a 32-byte key from the token
	hash := sha256.Sum256([]byte(*cd.token))
	return hash[:], nil
}

// EncryptData encrypts the given data using AES-GCM encryption
func (cd *CryptData) EncryptData(data interface{}) ([]byte, error) {
	if cd.EncryptionKey == nil {
		return nil, fmt.Errorf(utils.EncryptionError, "cannot be nil")
	}

	if len(cd.EncryptionKey) == 0 {
		return nil, fmt.Errorf(utils.EncryptionError, "cannot be empty")
	}

	// Convert data to JSON bytes
	var plaintext []byte
	var err error

	switch v := data.(type) {
	case string:
		plaintext = []byte(v)
	case []byte:
		plaintext = v
	default:
		plaintext, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
	}

	// Create AES cipher
	block, err := aes.NewCipher(cd.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext
	encryptedData := append(nonce, ciphertext...)

	// Return hex encoded data
	return []byte(hex.EncodeToString(encryptedData)), nil
}

// DecryptData decrypts the given data using AES-GCM decryption
func (cd *CryptData) DecryptData(data []byte) ([]byte, error) {
	if cd.EncryptionKey == nil {
		return nil, fmt.Errorf(utils.DecryptionError, "cannot be nil")
	}

	if len(cd.EncryptionKey) == 0 {
		return nil, fmt.Errorf(utils.DecryptionError, "cannot be empty")
	}

	// Decode hex string
	encryptedData, err := hex.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex data: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(cd.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Check minimum length
	if len(encryptedData) < gcm.NonceSize() {
		return nil, errors.New("encrypted data too short")
	}

	// Extract nonce and ciphertext
	nonce := encryptedData[:gcm.NonceSize()]
	ciphertext := encryptedData[gcm.NonceSize():]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// ValidateKey checks if the provided key is valid for AES encryption
func (cd *CryptData) ValidateKey(key []byte) bool {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return false
	}

	// Try to create a cipher to validate the key
	_, err := aes.NewCipher(key)
	return err == nil
}

func (cd *CryptData) GenerateHash(data interface{}) string {
	b := sha256.Sum256([]byte(*cd.token))

	fmt.Println("Generating hash for data:", data)
	fmt.Println("convert to data :", b)

	return hex.EncodeToString(b[:])
}
