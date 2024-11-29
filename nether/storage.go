package nether

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Storage struct {
	CountEmbeddings uint8
	CountImages     uint8
	Embeddings      []Embedding
	Images          []Image
}

func (s *Storage) Serialize() []byte {
	buffer := new(bytes.Buffer)

	if err := binary.Write(buffer, binary.LittleEndian, s.CountEmbeddings); err != nil {
		fmt.Printf("erro ao serializar CountEmbeddings: %v", err)
		return nil
	}

	if err := binary.Write(buffer, binary.LittleEndian, s.CountImages); err != nil {
		fmt.Printf("erro ao serializar CountImages: %v", err)
		return nil
	}

	for _, embedding := range s.Embeddings {
		serializedEmbedding, err := embedding.Serialize()
		if err != nil {
			fmt.Printf("erro ao serializar embedding: %v", err)
			return nil
		}
		buffer.Write(serializedEmbedding)
	}

	for _, image := range s.Images {
		serializedImage, err := image.Serialize()
		if err != nil {
			fmt.Printf("erro ao serializar imagem: %v", err)
			return nil
		}
		buffer.Write(serializedImage)
	}

	return buffer.Bytes()
}

func (s *Storage) Deserialize(data []byte) error {
	buffer := bytes.NewReader(data)

	if err := binary.Read(buffer, binary.LittleEndian, &s.CountEmbeddings); err != nil {
		return fmt.Errorf("erro ao desserializar CountEmbeddings: %v", err)
	}

	if err := binary.Read(buffer, binary.LittleEndian, &s.CountImages); err != nil {
		return fmt.Errorf("erro ao desserializar CountImages: %v", err)
	}

	s.Embeddings = make([]Embedding, s.CountEmbeddings)
	for i := uint8(0); i < s.CountEmbeddings; i++ {
		var embedding Embedding
		embeddingData := make([]byte, EMBEDDING_SIZE)

		if _, err := buffer.Read(embeddingData); err != nil {
			return fmt.Errorf("erro ao ler dados do embedding: %v", err)
		}

		if err := embedding.Deserialize(embeddingData); err != nil {
			return fmt.Errorf("erro ao desserializar embedding: %v", err)
		}

		s.Embeddings[i] = embedding
	}

	s.Images = make([]Image, s.CountImages)
	for i := uint8(0); i < s.CountImages; i++ {
		var image Image
		imageData := make([]byte, IMAGE_SIZE)

		if _, err := buffer.Read(imageData); err != nil {
			return fmt.Errorf("erro ao ler dados da imagem: %v", err)
		}

		if err := image.Deserialize(imageData); err != nil {
			return fmt.Errorf("erro ao desserializar imagem: %v", err)
		}

		s.Images[i] = image
	}

	return nil
}

func NewStorage(e []Embedding, i []Image) *Storage {
	return &Storage{
		CountEmbeddings: uint8(len(e)),
		CountImages:     uint8(len(i)),
		Embeddings:      e,
		Images:          i,
	}
}
