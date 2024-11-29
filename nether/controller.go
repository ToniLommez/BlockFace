package nether

import (
	"fmt"
	"log"
	"math/rand"
	"os"
)

const (
	USERDATA_PATH string = "data/nether.conf"
)

var (
	userdata *UserData
	reader   *NetherReader
)

func Start() {
	initHandlers()
	initServer()
}

func StartLog() {
	// Cria ou abre o arquivo para log
	file, err := os.OpenFile(LOG_PATH, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Erro ao abrir arquivo de log: %v", err)
	}
	defer file.Close()

	// Define a saída do log para o arquivo
	log.SetOutput(file)

	// Exemplos de log
	log.Println("Iniciando o programa...")
}

func Register(password string) {
	userdata = register(password)
}

func GetUserdata() string {
	return fmt.Sprintf("%v", userdata)
}

func ResetUserdata() {
	userdata = &UserData{}
}

func LoadData(password string) (success bool) {
	success, userdata = LoadConfig(password)
	return
}

func NewBlockchain() {
	reader, _ = newBlockchain(userdata.Key)
}

func LoadBlockchain() {
	reader, _ = NewReader()
}

func WriteRandomBlock() {
	// Random Size
	numEmbeddings := 2 + rand.Intn(4)
	numImages := 1 + rand.Intn(3)

	// Random embeddings
	embeddings := make([]Embedding, numEmbeddings)
	for i := 0; i < numEmbeddings; i++ {
		var data [128]float64
		for j := 0; j < 128; j++ {
			data[j] = float64(j + 1)
		}

		tmp, _ := newEmbedding(data)
		embeddings[i] = *tmp
	}

	// Random Images
	images := make([]Image, numImages)
	for i := 0; i < numImages; i++ {
		img, err := generateRandomImage()
		if err != nil {
			// fmt.Printf("Erro ao gerar imagem aleatória: %v\n", err)
			return
		}
		images[i] = *img
	}

	WriteBlock(NewStorage(embeddings, images))
}

func WriteBlock(storage *Storage) {
	b, _ := NewBlock(reader.ReadLastBlock(), userdata.Key, *storage)
	reader.WriteBlock(b)
}

func PrintBlockchain() {
	fmt.Printf("Metadata:\n")
	fmt.Printf("%v\n", reader)
	reader.ReadGenesis()
	fmt.Printf("Genesis:\n")
	fmt.Printf("%v\n", reader.current)
	for reader.ReadNext() {
		fmt.Printf("Block [%03d]:\n", reader.current.Index)
		fmt.Printf("%s\n", reader.current)
	}
}
