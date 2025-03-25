package nether

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
)

const (
	EMBEDDING_SIZE = (128 * 8) + 32
)

type Embedding struct {
	id   [32]byte
	data [128]float64
}

func (e *Embedding) generateId() {
	embeddingStr := fmt.Sprintf("%v", e)
	e.id = sha256.Sum256([]byte(embeddingStr))
}

func (e *Embedding) Serialize() ([]byte, error) {
	buffer := new(bytes.Buffer)

	if err := binary.Write(buffer, binary.LittleEndian, e.id); err != nil {
		return nil, fmt.Errorf("erro ao serializar ID: %v", err)
	}

	if err := binary.Write(buffer, binary.LittleEndian, e.data); err != nil {
		return nil, fmt.Errorf("erro ao serializar data: %v", err)
	}

	return buffer.Bytes(), nil
}

func (e *Embedding) Deserialize(data []byte) error {
	buffer := bytes.NewReader(data)

	if err := binary.Read(buffer, binary.LittleEndian, &e.id); err != nil {
		return fmt.Errorf("erro ao desserializar ID: %v", err)
	}

	if err := binary.Read(buffer, binary.LittleEndian, &e.data); err != nil {
		return fmt.Errorf("erro ao desserializar data: %v", err)
	}

	return nil
}

func newEmbedding(data [128]float64) (*Embedding, error) {
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			value, _ := rand.Int(rand.Reader, big.NewInt(100))
			data[i] = float64(value.Int64())
		}
	}

	e := &Embedding{data: data}
	e.generateId()

	return e, nil
}
