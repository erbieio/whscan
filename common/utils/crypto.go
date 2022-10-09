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

// PubkeyToAddress public key to address
func PubkeyToAddress(p *secp256k1.PublicKey) types.Address {
	data := elliptic.Marshal(secp256k1.S256(), p.X(), p.Y())
	return types.Address("0x" + hex.EncodeToString(Keccak256(data[1:])[12:]))
}

// NewTx creates a transaction, nonce: hexadecimal string with prefix, gasPrice: decimal string
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

// SignTx sign the transaction
func SignTx(tx map[string]string, chainId *big.Int, prv *secp256k1.PrivateKey) (types.Data, error) {
	msg := rlp.Value{
		List: []rlp.Value{
			{String: tx["nonce"]},    //0:nonce
			{String: tx["gasPrice"]}, //1:gasPrice
			{String: tx["gas"]},      //2:gas
			{String: tx["to"]},       //3:to
			{String: tx["value"]},    //4:value
			{String: tx["data"]},     //5:data
			{String: "0x" + hex.EncodeToString(chainId.Bytes())}, //6:V, equal to ChainId before unsigned
			{String: "0x"}, //7:R
			{String: "0x"}, //8:S
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
	R, S, V, err := DecodeEIP155Sig(sig, chainId)
	if err != nil {
		return "", err
	}
	msg.List[6] = rlp.Value{String: "0x" + hex.EncodeToString(V.Bytes())}
	msg.List[7] = rlp.Value{String: "0x" + hex.EncodeToString(R.Bytes())}
	msg.List[8] = rlp.Value{String: "0x" + hex.EncodeToString(S.Bytes())}
	raw, err := msg.Encode()
	return types.Data(raw), err
}

func DecodeEIP155Sig(sig []byte, chainId *big.Int) (R, S, V *big.Int, err error) {
	if len(sig) != 65 {
		return nil, nil, nil, fmt.Errorf("sig length is not 65:%d", len(sig))
	}
	R, S, V = new(big.Int), new(big.Int), new(big.Int)
	R.SetBytes(sig[0:32])
	S.SetBytes(sig[32:64])
	if chainId.Sign() == 0 {
		// v + 27
		V.SetBytes([]byte{sig[64] + 27})
	} else {
		// v + chainId * 2 + 35
		V = V.Lsh(chainId, 1).Add(V, big.NewInt(int64(sig[64]+35)))
	}
	return
}

// Sign signed with the private key, the last bit of the result is v, the value is 0 or 1
func Sign(digestHash []byte, prv *secp256k1.PrivateKey) ([]byte, error) {
	if len(digestHash) != 32 {
		return nil, fmt.Errorf("hash requires 32 bytes: %d", len(digestHash))
	}
	sig := ecdsa.SignCompact(prv, digestHash, false)
	// Subtract 27 from v and put it at the end
	return append(sig[1:65], sig[0]-27), nil
}

// SigToPub signature recovery public key
func SigToPub(hash, sig []byte) (*secp256k1.PublicKey, error) {
	s, _, err := ecdsa.RecoverCompact(append([]byte{sig[64] + 27}, sig[:64]...), hash)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Keccak256Hash calculates Keccak256 and returns the hash
func Keccak256Hash(data ...[]byte) (h types.Hash) {
	return types.Hash(hex.EncodeToString(Keccak256(data...)))
}

// Keccak256 Calculate Keccak256 return byte array (32 bytes)
func Keccak256(data ...[]byte) (h []byte) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}

	return d.Sum(nil)
}

// HexToECDSA hexadecimal string restore private key object
func HexToECDSA(key string) (*secp256k1.PrivateKey, error) {
	b, err := hex.DecodeString(key)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, fmt.Errorf("invalid hex data for private key")
	}
	return secp256k1.PrivKeyFromBytes(b), nil
}

// RecoverAddress recovers the address from the signature
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
