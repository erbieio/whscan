package utils

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"

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
