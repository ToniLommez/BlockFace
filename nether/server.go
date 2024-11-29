package nether

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type RequestData struct {
	Embeddings [][]float64 `json:"embeddings"`
	Images     []string    `json:"images"`
}

var (
	mu sync.Mutex
)

func initServer() {
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

	var embeddings []Embedding
	for _, data := range requestData.Embeddings {
		if len(data) != 128 {
			http.Error(w, "Cada embedding deve conter exatamente 128 floats.", http.StatusBadRequest)
			return
		}

		var embeddingData [128]float64
		copy(embeddingData[:], data)

		embedding, err := newEmbedding(embeddingData)
		if err != nil {
			http.Error(w, "Erro ao criar embedding: "+err.Error(), http.StatusInternalServerError)
			return
		}
		embeddings = append(embeddings, *embedding)
	}

	var images []Image
	for _, encodedImage := range requestData.Images {
		img, err := newImage(encodedImage)
		if err != nil {
			http.Error(w, "Erro ao processar imagem em Base64: "+err.Error(), http.StatusBadRequest)
			return
		}
		images = append(images, *img)
	}

	mu.Lock()
	WriteBlock(NewStorage(embeddings, images))
	mu.Unlock()
}
