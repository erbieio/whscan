package utils

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

func Test(t *testing.T) {
	hexKey := "7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd"
	key, _ := crypto.HexToECDSA(hexKey)
	addr := crypto.PubkeyToAddress(key.PublicKey).Hex()
	data := []byte(time.Now().UTC().Format("20060102"))

	sig, _ := Sign(data, key)
	if !VerifyDateSig(sig, addr) {
		t.Error("签名生成与验证错误")
	}
}

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
