package nether

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	METADATA_PATH string = "data/nether.conf"
)

type Metadata struct {
	Hash Hash
	Key  Key
}

var metadata Metadata

func (m *Metadata) String() string {
	return fmt.Sprintf("Hash: %s\nPrivateKey: %s\nPublicKey: %s",
		hex.EncodeToString(m.Hash[:]),
		hex.EncodeToString(m.Key.Sk[:]),
		hex.EncodeToString(m.Key.Pk[:]))
}

func GetMetadata() *Metadata {
	return &metadata
}

func ResetMetadata() {
	metadata = Metadata{}
}

func Register(password string) {
	metadata = Metadata{
		Hash: HashPassword(password),
		Key:  *NewKey(),
	}

	SaveConfig()
}

// SaveConfig save metadatas on file using AES with GCM mode
func SaveConfig() {
	file, err := os.OpenFile(METADATA_PATH, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		panic(fmt.Errorf("cannot save configurations: %w", err))
	}
	defer file.Close()

	jsonData, _ := json.Marshal(metadata)

	// Create a new AES cipher using GCM mode
	block, _ := aes.NewCipher(metadata.Hash[:])
	aesGCM, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesGCM.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ciphertext := aesGCM.Seal(nil, nonce, jsonData, nil)

	file.Write(nonce)
	file.Write(ciphertext)
}

func LoadConfig(password string) bool {
	file, err := os.Open(METADATA_PATH)
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
		return false
	}

	// Recover metadata
	json.Unmarshal(plaintext, &metadata)
	return true
}
