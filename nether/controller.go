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
	n := 2 + rand.Intn(4)
	sl := make([]StorageLocation, n)
	for i := 0; i < n; i++ {
		sl[i] = *NewStorageLocation(userdata.Key.Pk, uint64(rand.Intn(101)))
	}
	WriteBlock(NewDataSet(sl))
}

func WriteBlock(data *DataSet) {
	b, _ := NewBlock(reader.ReadLastBlock(), userdata.Key, *data)
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
		fmt.Printf("%v\n", reader.current)
	}
}
