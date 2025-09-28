package sauri

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type Encryption struct {
	Key []byte
}

// Encrypt encrypts the plaintext using AES and returns the
// ciphertext as a base64 encoded string
func (e *Encryption) Encrypt(text string) (string, error) {

	plainText := []byte(text)

	//todo Key Initialization

	// Create a new AES cipher with the provided key
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err // Return an error if cipher creation fails
	}

	// Create a byte slice for the ciphertext, which is the size of the AES block plus the length of the plaintext
	ciphertext := make([]byte, aes.BlockSize+len(plainText))

	// Create an initialization vector (IV) from the first part of the ciphertext
	iv := ciphertext[:aes.BlockSize]

	// Fill the IV with random bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err // Return an error if IV generation fails
	}

	// Create a new CFB encrypter with the cipher block and IV
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt the plaintext by XORing it with the key stream
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plainText)

	// Return the ciphertext as a base64 encoded string
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt function decrypts the base64 encoded ciphertext using AES
// and returns the plaintext
func (e *Encryption) Decrypt(ciphertext string) (string, error) {
	// Decode the base64 encoded ciphertext
	ciphertextBytes, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err // Return an error if decoding fails
	}
	// Create a new AES cipher with the provided key
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err // Return an error if cipher creation fails
	}

	// Check if the ciphertext is at least as long as the AES block size
	if len(ciphertextBytes) < aes.BlockSize {
		return "", errors.New("ciphertext too short") // Return an error if the ciphertext is too short
	}

	// Extract the initialization vector (IV) from the ciphertext
	iv := ciphertextBytes[:aes.BlockSize]

	// Extract the actual ciphertext
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	// Create a new CFB decrypter with the cipher block and IV
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt the ciphertext by XORing it with the key stream
	stream.XORKeyStream(ciphertextBytes, ciphertextBytes)

	// Return the decrypted plaintext
	return string(ciphertextBytes), nil
}
