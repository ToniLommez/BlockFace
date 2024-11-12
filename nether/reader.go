package nether

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Base files path
const (
	BLOCKCHAIN_PATH string = "data/nether.chain"
	IMAGES_PATH     string = "data/nether.database"
	METADATA_SIZE   int64  = int64(8 + 8 + 8 + 8 + CIPHER_SIZE)
)

// NetherReader have the reader state of the blockchain
type NetherReader struct {
	file            *os.File
	current         *Block
	size            uint64
	localSize       uint64
	lastBlockIndex  uint64
	lastBlockOffset uint64
	firstBlockHash  Hash
}

// Close encapsula o fechamento do arquivo
func (r *NetherReader) Close() error {
	r.current = nil
	return r.file.Close()
}

// NewReader create a new reader for the blockchain
func NewReader() (*NetherReader, error) {
	file, err := os.OpenFile(BLOCKCHAIN_PATH, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	reader := &NetherReader{
		file:            file,
		current:         nil,
		size:            0,
		localSize:       0,
		lastBlockIndex:  0,
		lastBlockOffset: 0,
		firstBlockHash:  [CIPHER_SIZE]byte{},
	}

	reader.ReadMetadata()
	reader.ReadGenesis()

	return reader, nil
}

func (r *NetherReader) SkipMetadata() {
	r.file.Seek(METADATA_SIZE, io.SeekStart)
}

func (r *NetherReader) ReadMetadata() {
	r.file.Seek(0, io.SeekStart)
	binary.Read(r.file, binary.LittleEndian, &r.size)
	binary.Read(r.file, binary.LittleEndian, &r.localSize)
	binary.Read(r.file, binary.LittleEndian, &r.lastBlockIndex)
	binary.Read(r.file, binary.LittleEndian, &r.lastBlockOffset)
	binary.Read(r.file, binary.LittleEndian, r.firstBlockHash)
}

func (r *NetherReader) WriteMetadata() {
	r.file.Seek(0, io.SeekStart)
	binary.Write(r.file, binary.LittleEndian, r.size)
	binary.Write(r.file, binary.LittleEndian, r.localSize)
	binary.Write(r.file, binary.LittleEndian, r.lastBlockIndex)
	binary.Write(r.file, binary.LittleEndian, r.lastBlockOffset)
	binary.Write(r.file, binary.LittleEndian, r.firstBlockHash)
}

func (r *NetherReader) WriteGenesis(k Key, genesisHash Hash) {
	r.SkipMetadata()
	binary.Write(r.file, binary.LittleEndian, NewGenesis(k, genesisHash).Serialize())
	r.file.Sync()
}

func (r *NetherReader) ReadGenesis() {
	r.SkipMetadata()
	r.readBlock()
}

func (r *NetherReader) ReadNext() bool {
	if r.current.Index >= r.lastBlockIndex {
		return false
	}

	r.readBlock()
	return true
}

func (r *NetherReader) readBlock() {
	var blockSize uint64
	binary.Read(r.file, binary.LittleEndian, &blockSize)

	rawBlock := make([]byte, blockSize)
	binary.Read(r.file, binary.LittleEndian, rawBlock[8:])
	binary.LittleEndian.PutUint64(rawBlock[:8], blockSize)

	r.current = Deserialize(rawBlock)
}

func (r *NetherReader) ReadLastBlock() *Block {
	r.file.Seek(int64(r.lastBlockOffset), io.SeekStart)
	r.readBlock()
	return r.current
}

func (r *NetherReader) WriteBlock(b *Block) {
	offset, _ := r.file.Seek(0, io.SeekEnd)
	if err := binary.Write(r.file, binary.LittleEndian, b.Serialize()); err != nil {
		panic(err)
	}
	r.current = b
	r.size++
	r.localSize++
	r.lastBlockIndex = b.Index
	r.lastBlockOffset = uint64(offset)
	r.WriteMetadata()
	r.ReadLastBlock()
}

func newBlockchain(k Key) (*NetherReader, error) {
	file, err := os.OpenFile(BLOCKCHAIN_PATH, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new blockchain: %w", err)
	}

	var genesisHash Hash
	io.ReadFull(rand.Reader, genesisHash[:])

	r := &NetherReader{
		file:            file,
		current:         nil,
		size:            1,
		localSize:       1,
		lastBlockOffset: uint64(METADATA_SIZE),
		firstBlockHash:  genesisHash,
	}

	r.WriteMetadata()
	r.WriteGenesis(k, genesisHash)

	return r, nil
}
