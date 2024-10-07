package nether

import (
	"bytes"
	"encoding/binary"
)

const (
	STORAGE_LOCATION_SIZE = PUBLIC_KEY_SIZE + 8
)

type DataSet struct {
	Count  uint8
	Values []StorageLocation
}

type StorageLocation struct {
	Key     [PUBLIC_KEY_SIZE]byte
	Pointer uint64
}

func (d *DataSet) Serialize() []byte {
	var buf bytes.Buffer

	for _, v := range d.Values {
		binary.Write(&buf, binary.LittleEndian, v.Key)
		binary.Write(&buf, binary.LittleEndian, v.Pointer)
	}

	return buf.Bytes()
}

func (d *DataSet) Deserialize(data []byte) {
	buf := bytes.NewReader(data)

	dataSize := len(data)
	count := dataSize / STORAGE_LOCATION_SIZE
	if dataSize%STORAGE_LOCATION_SIZE != 0 {
		panic("Invalid data size")
	}

	d.Values = make([]StorageLocation, count)
	for i := 0; i < count; i++ {
		buf.Read(d.Values[i].Key[:])
		binary.Read(buf, binary.LittleEndian, &d.Values[i].Pointer)
	}

	d.Count = uint8(count)
}

func NewDataSet(sl []StorageLocation) *DataSet {
	return &DataSet{
		Count:  uint8(len(sl)),
		Values: sl,
	}
}

func NewStorageLocation(key [PUBLIC_KEY_SIZE]byte, pointer uint64) *StorageLocation {
	return &StorageLocation{
		Key:     key,
		Pointer: pointer,
	}
}
