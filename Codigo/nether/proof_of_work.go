package nether

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math/bits"
	"math/rand"
	"sync"
)

var (
	cancelFunc      context.CancelFunc
	THREADS         = 16
	LIMIT           = 100_000
	STOP_PROCESSING = false
	characters      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// Gera uma string incremental com base em um contador
func generateNonce(base int, counter int) string {
	nonce := ""
	for counter > 0 {
		nonce = string(characters[counter%base]) + nonce
		counter /= base
	}
	return nonce
}

func mine(message []byte, zeroes int, tid int, randomState int, ctx context.Context, wg *sync.WaitGroup, results chan<- string) {
	defer wg.Done()

	start := LIMIT*tid + randomState
	step := LIMIT * THREADS
	var buf bytes.Buffer

	for counter := start; ; counter++ {
		select {
		case <-ctx.Done():
			return
		default:
			// Gerando o nonce como string
			nonceStr := generateNonce(len(characters), counter)

			// Escrevendo o nonce no buffer
			buf.Reset()
			buf.WriteString(nonceStr)

			// Gerando o hash
			sha := sha256.Sum256(append(message, buf.Bytes()...))

			// Verificando o número de zeros à esquerda
			if leadingZeroBits(sha) >= zeroes {
				results <- nonceStr
				return
			}

			// Atualizando os limites
			if counter-start >= LIMIT {
				start += step
				counter = start
			}

			if STOP_PROCESSING {
				return
			}
		}
	}
}

func proof_of_work(zeroes int, message []byte) (string, bool) {
	var wg sync.WaitGroup
	results := make(chan string, THREADS)

	ctx, cancel := context.WithCancel(context.Background())
	cancelFunc = cancel

	for i := 0; i < THREADS; i++ {
		wg.Add(1)
		go mine(message, zeroes, i, rand.Intn(1_000_000_000), ctx, &wg, results)
	}

	found := false
	var nonce string
	select {
	case nonce = <-results:
		found = true
		cancel()
	case <-ctx.Done():
		nonce = ""
	}

	wg.Wait()
	close(results)

	return nonce, found
}

func validateProof(message []byte, nonce string, zeroes int) bool {
	var buf bytes.Buffer
	buf.Reset()
	buf.WriteString(nonce)

	sha := sha256.Sum256(append(message, buf.Bytes()...))

	return leadingZeroBits(sha) >= zeroes
}

func getHash(message []byte, nonce string) string {
	return fmt.Sprintf("%x", sha256.Sum256(append(message, []byte(nonce)...)))
}

func leadingZeroBits(hash [32]byte) int {
	count := 0
	for _, b := range hash {
		if b == 0 {
			count += 8
		} else {
			count += bits.LeadingZeros8(b)
			break
		}
	}
	return count
}
