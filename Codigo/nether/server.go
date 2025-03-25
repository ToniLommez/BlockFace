package nether

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type RequestData struct {
	Embeddings []float64 `json:"embeddings"`
	ImagePaths string    `json:"image_path"`
}

func InitServer() {
	http.HandleFunc("/add", addToBlockchainHandler)

	fmt.Println("Servidor iniciado em http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Erro ao iniciar o servidor: %v\n", err)
	}
}

func cors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func addToBlockchainHandler(w http.ResponseWriter, r *http.Request) {
	cors(w, r)

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erro ao ler o corpo da requisição", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var requestData RequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Erro ao parsear JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validar embedding
	if len(requestData.Embeddings) != 128 {
		http.Error(w, "O embedding deve conter exatamente 128 floats.", http.StatusBadRequest)
		return
	}

	var embeddingData [128]float64
	copy(embeddingData[:], requestData.Embeddings)

	embedding, err := newEmbedding(embeddingData)
	if err != nil {
		http.Error(w, "Erro ao criar embedding: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validar caminho da imagem
	if _, err := os.Stat(requestData.ImagePaths); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("O arquivo de imagem não existe no caminho especificado: %s", requestData.ImagePaths), http.StatusBadRequest)
		return
	}

	image := Image{data: requestData.ImagePaths}

	// Adicionar ao blockchain
	WriteBlock(NewStorage(*embedding, image))
	fmt.Printf("Novo rosto adicionado a blockchain\n")

	// Responder ao cliente
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Dados adicionados ao blockchain com sucesso."))
}
