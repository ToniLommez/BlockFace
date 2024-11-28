package nether

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/bits"
	"math/rand"
	"sync"
)

var (
	cancelFunc      context.CancelFunc
	THREADS         = 8
	LIMIT           = 100_000
	STOP_PROCESSING = false
)

func mine(message []byte, zeroes int, tid int, randomState int, ctx context.Context, wg *sync.WaitGroup, results chan<- int) {
	defer wg.Done()

	// Inicialização dos limites com randomState
	start := LIMIT*tid + randomState
	step := LIMIT * THREADS
	var buf bytes.Buffer

	for nonce := start; ; nonce++ {
		select {
		case <-ctx.Done():
			return
		default:
			// Escrevendo o nonce no buffer
			buf.Reset()
			binary.Write(&buf, binary.BigEndian, int64(nonce))

			// Gerando o hash
			sha := sha256.Sum256(append(message, buf.Bytes()...))

			// Verificando o número de zeros à esquerda
			if bits.LeadingZeros64(binary.BigEndian.Uint64(sha[:])) >= zeroes {
				results <- nonce
				return
			}

			// Atualizando os limites
			if nonce-start >= LIMIT {
				start += step
				nonce = start
			}

			if STOP_PROCESSING {
				return
			}
		}
	}
}

func proof_of_work(zeroes int, message []byte) (int, bool) {
	var wg sync.WaitGroup
	results := make(chan int, THREADS)

	ctx, cancel := context.WithCancel(context.Background())
	cancelFunc = cancel

	for i := 0; i < THREADS; i++ {
		wg.Add(1)
		go mine(message, zeroes, i, rand.Intn(1_000_000_000), ctx, &wg, results)
	}

	found := false
	var nonce int
	select {
	case nonce = <-results:
		found = true
		cancel()
	case <-ctx.Done():
		nonce = 0
	}

	wg.Wait()
	close(results)

	return nonce, found
}

func validateProof(message []byte, nonce int, zeroes int) bool {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, int64(nonce))

	sha := sha256.Sum256(append(message, buf.Bytes()...))

	return bits.LeadingZeros64(binary.BigEndian.Uint64(sha[:])) >= zeroes
}

func getHash(message []byte, nonce int) string {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, int64(nonce))

	sha := sha256.Sum256(append(message, buf.Bytes()...))

	return fmt.Sprintf("%x", sha)
}
