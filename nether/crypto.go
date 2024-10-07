package nether

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
)

const (
	CIPHER_SIZE     = 32
	SIGNATURE_SIZE  = 64
	SECRET_KEY_SIZE = 32
	PUBLIC_KEY_SIZE = 64
)

type PrivateKey = [SECRET_KEY_SIZE]byte
type PublicKey = [PUBLIC_KEY_SIZE]byte
type Signature = [SIGNATURE_SIZE]byte
type Hash = [CIPHER_SIZE]byte
type Key struct {
	Sk PrivateKey
	Pk PublicKey
}

func NewKey() *Key {
	k := &Key{PrivateKey{}, PublicKey{}}
	var privateKey *ecdsa.PrivateKey
	var err error

	if privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		panic(fmt.Sprintf("falha ao gerar o par de chaves: %v", err))
	}

	copy(k.Sk[:], privateKey.D.Bytes())
	copy(k.Pk[:32], privateKey.PublicKey.X.Bytes())
	copy(k.Pk[32:], privateKey.PublicKey.Y.Bytes())

	return k
}

func BytesToEcdsaPrivateKey(keyBytes PrivateKey) *ecdsa.PrivateKey {
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
		D: new(big.Int).SetBytes(keyBytes[:]),
	}
}

func BytesToEcdsaPublicKey(keyBytes PublicKey) *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(keyBytes[:32]),
		Y:     new(big.Int).SetBytes(keyBytes[32:]),
	}
}

func CipherPointer(k Key, pointer uint64) *StorageLocation {
	v1 := binary.LittleEndian.Uint64(k.Sk[0:8])
	v2 := binary.LittleEndian.Uint64(k.Sk[8:16])
	v3 := binary.LittleEndian.Uint64(k.Sk[16:24])
	v4 := binary.LittleEndian.Uint64(k.Sk[24:32])

	v := v1 ^ v2 ^ v3 ^ v4

	cipher := pointer ^ v

	return NewStorageLocation(k.Pk, cipher)
}

func HashPassword(password string) [32]byte {
	return sha256.Sum256([]byte(password))
}
