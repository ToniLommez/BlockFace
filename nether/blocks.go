package nether

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"time"
)

type Block struct {
	BlockSize uint64
	Index     uint64
	Timestamp uint64
	PrevHash  Hash
	Hash      Hash
	Signature Signature
	PubKey    PublicKey
	Storage   Storage
}

func (b *Block) calculateHash() {
	record := make([]byte, 0)
	index := make([]byte, 8)
	timestamp := make([]byte, 8)

	binary.LittleEndian.PutUint64(index, b.Index)
	binary.LittleEndian.PutUint64(timestamp, b.Timestamp)

	record = append(record, index...)
	record = append(record, timestamp...)
	record = append(record, b.PrevHash[:]...)

	serializedEmbedding, _ := b.Storage.Embedding.Serialize()
	record = append(record, serializedEmbedding...)

	serializedImage, _ := b.Storage.Image.Serialize()
	record = append(record, serializedImage...)

	b.Hash = sha256.Sum256(record)
}

func (b *Block) sign(Sk PrivateKey) (err error) {
	r, s, err := ecdsa.Sign(rand.Reader, BytesToEcdsaPrivateKey(Sk), b.Hash[:])
	if err != nil {
		return err
	}

	copy(b.Signature[:32], r.Bytes())
	copy(b.Signature[32:], s.Bytes())

	return nil
}

func (b *Block) Verify(Pk PublicKey) bool {
	var r, s big.Int
	r.SetBytes(b.Signature[:32])
	s.SetBytes(b.Signature[32:])

	return ecdsa.Verify(BytesToEcdsaPublicKey(Pk), b.Hash[:], &r, &s)
}

func (b *Block) computeSize() {
	blockSize := 8
	index := 8
	timestamp := 8
	prevHash := CIPHER_SIZE
	hash := CIPHER_SIZE
	signature := SIGNATURE_SIZE
	pubKey := PUBLIC_KEY_SIZE

	embeddingsSize := EMBEDDING_SIZE
	imagesSize := 200
	storageSize := embeddingsSize + imagesSize

	b.BlockSize = uint64(blockSize + index + timestamp + prevHash + hash + signature + pubKey + storageSize)
}

func (b *Block) Serialize() []byte {
	var buf bytes.Buffer

	// Block metadata
	binary.Write(&buf, binary.LittleEndian, b.BlockSize)
	binary.Write(&buf, binary.LittleEndian, b.Index)
	binary.Write(&buf, binary.LittleEndian, b.Timestamp)

	// Fixed size arrays
	buf.Write(b.PrevHash[:])
	buf.Write(b.Hash[:])
	buf.Write(b.Signature[:])
	buf.Write(b.PubKey[:])

	// Storage
	buf.Write(b.Storage.Serialize())

	return buf.Bytes()
}

func Deserialize(data []byte) *Block {
	b := Block{}
	buf := bytes.NewReader(data)

	// Block metadata
	binary.Read(buf, binary.LittleEndian, &b.BlockSize)
	binary.Read(buf, binary.LittleEndian, &b.Index)
	binary.Read(buf, binary.LittleEndian, &b.Timestamp)

	// Fixed size arrays
	buf.Read(b.PrevHash[:])
	buf.Read(b.Hash[:])
	buf.Read(b.Signature[:])
	buf.Read(b.PubKey[:])

	// Storage
	b.Storage.Deserialize(data[len(data)-buf.Len():])

	return &b
}

func NewBlock(oldBlock *Block, k Key, store Storage) (*Block, error) {
	newBlock := &Block{
		BlockSize: 0,
		Index:     oldBlock.Index + 1,
		Timestamp: uint64(time.Now().Unix()),
		PrevHash:  oldBlock.Hash,
		PubKey:    k.Pk,
		Storage:   store,
	}

	newBlock.calculateHash()

	if newBlock.sign(k.Sk) != nil {
		panic("Cannot Sign")
	}

	newBlock.computeSize()

	return newBlock, nil
}

func NewGenesis(k Key, genesisHash Hash) *Block {
	genesis := &Block{
		BlockSize: 0,
		Index:     0,
		Timestamp: uint64(time.Now().Unix()),
		PrevHash:  genesisHash,
		PubKey:    k.Pk,
		Storage:   Storage{},
	}

	genesis.calculateHash()

	if genesis.sign(k.Sk) != nil {
		panic("Cannot Sign")
	}

	genesis.computeSize()

	return genesis
}
