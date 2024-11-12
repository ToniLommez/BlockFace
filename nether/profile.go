package nether

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	LOG_PATH string = "data/program.log"
)

type UserData struct {
	Hash Hash
	Key  Key
}

func register(password string) *UserData {
	m := &UserData{
		Hash: HashPassword(password),
		Key:  *NewKey(),
	}

	SaveConfig(m)

	return m
}

// SaveConfig save metadatas on file using AES with GCM mode
func SaveConfig(m *UserData) {
	file, err := os.OpenFile(USERDATA_PATH, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		panic(fmt.Errorf("cannot save configurations: %w", err))
	}
	defer file.Close()

	jsonData, _ := json.Marshal(*m)

	// Create a new AES cipher using GCM mode
	block, _ := aes.NewCipher(m.Hash[:])
	aesGCM, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesGCM.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ciphertext := aesGCM.Seal(nil, nonce, jsonData, nil)

	file.Write(nonce)
	file.Write(ciphertext)
}

func LoadConfig(password string) (bool, *UserData) {
	file, err := os.Open(USERDATA_PATH)
	if err != nil {
		panic(fmt.Errorf("erro ao abrir arquivo: %w", err))
	}
	defer file.Close()

	// Decipher text using password
	hash := HashPassword(password)
	block, _ := aes.NewCipher(hash[:])
	aesGCM, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesGCM.NonceSize())
	io.ReadFull(file, nonce)
	ciphertext, _ := io.ReadAll(file)
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return false, nil
	}

	// Recover metadata
	m := &UserData{}
	json.Unmarshal(plaintext, &m)
	return true, m
}
