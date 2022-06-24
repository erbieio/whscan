package utils

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/INFURA/go-ethlibs/rlp"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"
	"server/common/types"
)

// PubkeyToAddress 公钥转地址
func PubkeyToAddress(p *secp256k1.PublicKey) types.Address {
	data := elliptic.Marshal(secp256k1.S256(), p.X(), p.Y())
	return types.Address("0x" + hex.EncodeToString(Keccak256(data[1:])[12:]))
}

// NewTx 创建交易，nonce：带前缀16进制字符串，gasPrice：10进制字符串
func NewTx(nonce uint64, to types.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) map[string]string {
	tx := map[string]string{}
	tx["nonce"] = "0x" + hex.EncodeToString(big.NewInt(int64(nonce)).Bytes())
	tx["gasPrice"] = "0x" + hex.EncodeToString(gasPrice.Bytes())
	tx["gas"] = "0x" + hex.EncodeToString(big.NewInt(int64(gasLimit)).Bytes())
	tx["to"] = string(to)
	tx["value"] = "0x" + hex.EncodeToString(amount.Bytes())
	tx["data"] = "0x" + hex.EncodeToString(data)
	return tx
}

// SignTx 签名交易
func SignTx(tx map[string]string, chainId *big.Int, prv *secp256k1.PrivateKey) (types.Data, error) {
	msg := rlp.Value{
		List: []rlp.Value{
			{String: tx["nonce"]},    //0,nonce
			{String: tx["gasPrice"]}, //1,gasPrice
			{String: tx["gas"]},      //2,gas
			{String: tx["to"]},       //3,to
			{String: tx["value"]},    //4,value
			{String: tx["data"]},     //5,data
			{String: "0x" + hex.EncodeToString(chainId.Bytes())}, //6,V,未签名之前等于ChainId
			{String: "0x"}, //7,R
			{String: "0x"}, //8,V
		},
	}
	hash, err := msg.HashToBytes()
	if err != nil {
		return "", err
	}
	sig, err := Sign(hash, prv)
	if err != nil {
		return "", err
	}

	if tx["chainId"] == "0x" {
		// v + 27
		msg.List[6] = rlp.Value{String: "0x" + hex.EncodeToString([]byte{sig[64] + 27})}
	} else {
		// v + chainId * 2 + 35
		v := new(big.Int).Set(chainId)
		v = v.Lsh(v, 1)
		v = v.Add(v, big.NewInt(int64(sig[64]+35)))
		msg.List[6] = rlp.Value{String: "0x" + hex.EncodeToString(v.Bytes())}
	}
	msg.List[7] = rlp.Value{String: "0x" + hex.EncodeToString(sig[0:32])}
	msg.List[8] = rlp.Value{String: "0x" + hex.EncodeToString(sig[32:64])}
	raw, err := msg.Encode()
	return types.Data(raw), err
}

// Sign 用私钥签名，结果最后一位是v，值为0或1
func Sign(digestHash []byte, prv *secp256k1.PrivateKey) ([]byte, error) {
	if len(digestHash) != 32 {
		return nil, fmt.Errorf("哈希需要32字节：%d", len(digestHash))
	}
	sig := ecdsa.SignCompact(prv, digestHash, false)
	// 将v减去27并放到最后
	return append(sig[1:65], sig[0]-27), nil
}

// SigToPub 签名恢复公钥
func SigToPub(hash, sig []byte) (*secp256k1.PublicKey, error) {
	s, _, err := ecdsa.RecoverCompact(append([]byte{sig[64] + 27}, sig[:64]...), hash)
	if err != nil {
		return nil, err
	}

	return s, nil
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
func HexToECDSA(key string) (*secp256k1.PrivateKey, error) {
	b, err := hex.DecodeString(key)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, fmt.Errorf("invalid hex data for private key")
	}
	return secp256k1.PrivKeyFromBytes(b), nil
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
