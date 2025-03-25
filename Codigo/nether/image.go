package nether

import (
	"bytes"
	"fmt"
)

type Image struct {
	data string
}

func (img *Image) Serialize() ([]byte, error) {
	if len(img.data) > 200 {
		return nil, fmt.Errorf("string excede o tamanho máximo permitido de 200 caracteres")
	}

	paddedData := fmt.Sprintf("%-200s", img.data)
	dataBytes := []byte(paddedData)

	buffer := new(bytes.Buffer)

	if _, err := buffer.Write(dataBytes); err != nil {
		return nil, fmt.Errorf("erro ao escrever dados da string: %v", err)
	}

	return buffer.Bytes(), nil
}

func (img *Image) Deserialize(data []byte) error {
	if len(data) != 200 {
		return fmt.Errorf("tamanho dos dados inválido: esperado 200 bytes, recebido %d bytes", len(data))
	}

	img.data = string(bytes.TrimRight(data, " "))

	return nil
}
