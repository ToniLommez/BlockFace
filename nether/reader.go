package nether

import (
	"encoding/binary"
	"fmt"
	"os"
)

// Base files path
const (
	BLOCKCHAIN_PATH string = "data/nether.chain"
	IMAGES_PATH     string = "data/nether.database"
)

// NetherReader have the reader state of the blockchain
type NetherReader struct {
	file            *os.File
	current         *Block
	size            uint64
	localSize       uint64
	lastBlockOffset uint64
	firstBlockHash  Hash
}

// Close encapsula o fechamento do arquivo
func (r *NetherReader) Close() error {
	r.current = nil
	return r.file.Close()
}

// NewReader create a new reader for the blockchain
func NewReader(filePath string) (*NetherReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	reader := &NetherReader{
		file:            file,
		current:         nil,
		size:            0,
		localSize:       0,
		lastBlockOffset: 0,
		firstBlockHash:  [CIPHER_SIZE]byte{},
	}

	reader.readMetadata()
	reader.readGenesis()

	return reader, nil
}

func (r *NetherReader) readMetadata() {
	binary.Read(r.file, binary.LittleEndian, &r.size)
	binary.Read(r.file, binary.LittleEndian, &r.localSize)
	binary.Read(r.file, binary.LittleEndian, &r.lastBlockOffset)
	binary.Read(r.file, binary.LittleEndian, r.firstBlockHash)
}

func (r *NetherReader) writeMetadata() {
	binary.Write(r.file, binary.LittleEndian, r.size)
	binary.Write(r.file, binary.LittleEndian, r.localSize)
	binary.Write(r.file, binary.LittleEndian, r.lastBlockOffset)
	binary.Write(r.file, binary.LittleEndian, r.firstBlockHash)
}

func (r *NetherReader) readGenesis() {
	var blockSize uint64
	binary.Read(r.file, binary.LittleEndian, blockSize)

	rawBlock := make([]byte, blockSize)
	binary.LittleEndian.PutUint64(rawBlock[:8], blockSize)
	binary.Read(r.file, binary.LittleEndian, rawBlock[8:])

	r.current = Deserialize(rawBlock)
}

func (r *NetherReader) writeGenesis(k Key) {
	binary.Write(r.file, binary.LittleEndian, NewGenesis(k).Serialize())
}

// NewBlockchain creates a new blockchain and initializes it
func NewBlockchain(k Key) (*NetherReader, error) {
	file, err := os.OpenFile(BLOCKCHAIN_PATH, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new blockchain: %w", err)
	}

	r := &NetherReader{
		file:            file,
		current:         nil,
		size:            1,
		localSize:       1,
		lastBlockOffset: 8 + 8 + 8 + CIPHER_SIZE,
		firstBlockHash:  [CIPHER_SIZE]byte{},
	}

	r.writeMetadata()
	r.writeGenesis(k)

	return r, nil
}
