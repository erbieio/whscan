package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/sha3"
	"server/common/types"
)

// PubkeyToAddress 公钥转地址
func PubkeyToAddress(p *ecdsa.PublicKey) types.Address {
	data := elliptic.Marshal(btcec.S256(), p.X, p.Y)
	return types.Address("0x" + hex.EncodeToString(Keccak256(data[1:])[12:]))
}

// Sign 用私钥签名
func Sign(digestHash []byte, prv *btcec.PrivateKey) (sig []byte, err error) {
	if len(digestHash) != 32 {
		return nil, fmt.Errorf("哈希需要32字节：%d", len(digestHash))
	}
	s, err := prv.Sign(digestHash)
	if err != nil {
		return nil, err
	}
	// todo check
	return s.Serialize(), err
}

// SigToPub 签名恢复公钥
func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	s, _, err := btcec.RecoverCompact(btcec.S256(), sig, hash)
	if err != nil {
		return nil, err
	}

	return s.ToECDSA(), nil
}

// Keccak256Hash 计算Keccak256，返回哈希
func Keccak256Hash(data ...[]byte) (h types.Hash) {
	return types.Hash(hex.EncodeToString(Keccak256(data...)))
}

// Keccak256 计算Keccak256返回字节数组（32字节）
func Keccak256(data ...[]byte) (h []byte) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}

	return d.Sum(nil)
}

// HexToECDSA 16进制字符串还原私钥对象
func HexToECDSA(key string) (*btcec.PrivateKey, error) {
	b, err := hex.DecodeString(key)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, fmt.Errorf("invalid hex data for private key")
	}
	prv, _ := btcec.PrivKeyFromBytes(btcec.S256(), b)
	return prv, nil
}

// RecoverAddress 从签名恢复地址
func RecoverAddress(msg string, hexSig string) (types.Address, error) {
	sig, _ := hex.DecodeString(hexSig[2:])
	if len(sig) != 65 {
		return "", fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return "", fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[64] -= 27
	msg = fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
	hash := Keccak256([]byte(msg))
	rpk, err := SigToPub(hash, sig)
	if err != nil {
		return "", err
	}
	return PubkeyToAddress(rpk), nil
}
