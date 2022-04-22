package utils

import (
	"math/big"
	"testing"
)

func TestBigToAddress(t *testing.T) {
	a := big.NewInt(1)
	t.Log(BigToAddress(a))
	t.Log(BigToAddress(nil))
	a.SetString("bbbbbbbbbbbbbbbbbbbbbbbbaaaaaaaaaaaaaaaaaaaaaaaa", 16)
	t.Log(BigToAddress(a))
}

func TestKeccak256Hash(t *testing.T) {
	t.Log(Keccak256Hash([]byte("hello")))
}

func TestPubkeyToAddress(t *testing.T) {
	key, _ := HexToECDSA("7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd")
	t.Log(PubkeyToAddress(key.PublicKey))
	// 0x394586580ff4170c8a0244837202cbabe9070f66
}
