package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "errors"
    "io"
    "os"
)

const (
    nonceSize = 12 // GCM standard
)

func GenerateEncryptionKey() ([]byte, error) {
    key := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, key); err != nil {
        return nil, err
    }
    return key, nil
}

func EncryptFile(sourceFile, destFile string, key []byte) error {
    plaintext, err := os.ReadFile(sourceFile)
    if (err != nil) {
        return err
    }

    block, err := aes.NewCipher(key)
    if (err != nil) {
        return err
    }

    gcm, err := cipher.NewGCM(block)
    if (err != nil) {
        return err
    }

    nonce := make([]byte, nonceSize)
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return err
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

    return os.WriteFile(destFile, ciphertext, 0600)
}

func DecryptFile(sourceFile, destFile string, key []byte) error {
    ciphertext, err := os.ReadFile(sourceFile)
    if (err != nil) {
        return err
    }

    if len(ciphertext) < nonceSize {
        return errors.New("invalid encrypted file: too short")
    }

    block, err := aes.NewCipher(key)
    if (err != nil) {
        return err
    }

    gcm, err := cipher.NewGCM(block)
    if (err != nil) {
        return err
    }

    nonce := ciphertext[:nonceSize]
    ciphertext = ciphertext[nonceSize:]

    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if (err != nil) {
        return err
    }

    return os.WriteFile(destFile, plaintext, 0600)
}
