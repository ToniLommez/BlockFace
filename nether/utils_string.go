package nether

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func (b *Block) String() string {
	return fmt.Sprintf(
		"\tBlockSize: \t%v\n\tIndex: \t\t%v\n\tTimestamp: \t%v\n\tPrevHash: \t%s\n\tHash: \t\t%s\n\tSignature: \t%s\n\tPubKey: \t%s\n\tStorage: \t%v",
		b.BlockSize, b.Index, time.Unix(int64(b.Timestamp), 0).Format("02-01-2006 15:04:05"),
		base64.StdEncoding.EncodeToString(b.PrevHash[:]), base64.StdEncoding.EncodeToString(b.Hash[:]),
		base64.StdEncoding.EncodeToString(b.Signature[:]), base64.StdEncoding.EncodeToString(b.PubKey[:]),
		b.Storage.String())
}

func (m *UserData) String() string {
	return fmt.Sprintf("Hash: %s\nPrivateKey: %s\nPublicKey: %s",
		base64.StdEncoding.EncodeToString(m.Hash[:]),
		base64.StdEncoding.EncodeToString(m.Key.Sk[:]),
		base64.StdEncoding.EncodeToString(m.Key.Pk[:]))
}

func (k *Key) String() string {
	return fmt.Sprintf("Key (Sk: %s, Pk: %s)", k.Sk, k.Pk)
}

func (r NetherReader) String() string {
	fileInfo := "nil"
	if r.file != nil {
		fileInfo = r.file.Name()
	}

	blockInfo := "nil"
	if r.current != nil {
		blockInfo = fmt.Sprintf("Index: %d", r.current.Index)
	}

	return fmt.Sprintf(
		"\tFile: %s\n\tCurrent Block: %s\n\tSize: %d\n\tLocal Size: %d\n\tLast Block Index: %d\n\tLast Block Offset: %x\n\tFirst Block Hash: %s",
		fileInfo, blockInfo, r.size, r.localSize, r.lastBlockIndex, r.lastBlockOffset, base64.StdEncoding.EncodeToString(r.firstBlockHash[:]),
	)
}

func (e *Embedding) String() string {
	intRepresentation := make([]int, len(e.data))
	for i, v := range e.data {
		intRepresentation[i] = int(v)
	}

	return fmt.Sprintf(
		"(ID: %s, Data: %v)",
		base64.StdEncoding.EncodeToString(e.id[:]),
		intRepresentation[:5],
	)
}

func (s *Storage) String() string {
	return fmt.Sprintf(
		"Storage (\n\tEmbedding: %s\n\tImage: %s\n)",
		s.Embedding.String(),
		s.Image.data,
	)
}

func EncodePublicKey(src [64]byte) string {
	return base64.StdEncoding.EncodeToString(src[:])
}

func randomString(minLength, maxLength int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	length := rand.Intn(maxLength-minLength+1) + minLength

	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}

func chooseRandom(strs []string) (string, error) {
	if len(strs) == 0 {
		return "", fmt.Errorf("o slice está vazio, não é possível sortear um líder")
	}

	index := rand.Intn(len(strs))
	return strs[index], nil
}
