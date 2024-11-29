package nether

import (
	"bytes"
	"fmt"
)

type Storage struct {
	Embedding Embedding
	Image     Image
}

func (s *Storage) Serialize() []byte {
	buffer := new(bytes.Buffer)

	serializedEmbedding, err := s.Embedding.Serialize()
	if err != nil {
		fmt.Printf("erro ao serializar embedding: %v", err)
		return nil
	}
	buffer.Write(serializedEmbedding)

	serializedImage, err := s.Image.Serialize()
	if err != nil {
		fmt.Printf("erro ao serializar imagem: %v", err)
		return nil
	}
	buffer.Write(serializedImage)

	return buffer.Bytes()
}

func (s *Storage) Deserialize(data []byte) error {
	buffer := bytes.NewReader(data)

	embeddingData := make([]byte, EMBEDDING_SIZE)
	if _, err := buffer.Read(embeddingData); err != nil {
		return fmt.Errorf("erro ao ler dados do embedding: %v", err)
	}
	if err := s.Embedding.Deserialize(embeddingData); err != nil {
		return fmt.Errorf("erro ao desserializar embedding: %v", err)
	}

	imageData := make([]byte, 200)
	if _, err := buffer.Read(imageData); err != nil {
		return fmt.Errorf("erro ao ler dados da imagem: %v", err)
	}
	if err := s.Image.Deserialize(imageData); err != nil {
		return fmt.Errorf("erro ao desserializar imagem: %v", err)
	}

	return nil
}

func NewStorage(e Embedding, i Image) *Storage {
	return &Storage{
		Embedding: e,
		Image:     i,
	}
}
