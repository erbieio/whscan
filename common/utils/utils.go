package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/sha3"
	"server/common/types"
)

func ToBytes(v string) []byte {
	var bigTemp big.Int
	bigTemp.SetString(v, 0)
	return bigTemp.Bytes()
}

func HexToBigInt(hex string) types.BigInt {
	b := new(big.Int)
	b.UnmarshalText([]byte(hex))
	return types.BigInt(b.String())
}

func HexToAddress(hex string) (types.Address, error) {
	if len(hex) != 42 {
		return "", errors.New("长度不是42")
	}
	if hex[0] != '0' || (hex[1] != 'x' && hex[1] != 'X') {
		return "", errors.New("前缀不是0x")
	}
	for _, c := range hex[2:] {
		if ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F') {
			continue
		}
		return "", errors.New("非法字符:" + string(c))
	}
	return types.Address(strings.ToLower(hex)), nil
}

func BigToAddress(big *big.Int) types.Address {
	addr := "0000000000000000000000000000000000000000"
	if big != nil {
		addr += big.Text(16)
	}
	return types.Address("0x" + addr[len(addr)-40:])
}

func Keccak256Hash(data ...[]byte) (h types.Hash) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}

	return types.Hash("0x" + hex.EncodeToString(d.Sum(nil)))
}
func HexToECDSA(key string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(key)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, errors.New("invalid hex data for private key")
	}
	return secp256k1.PrivKeyFromBytes(b).ToECDSA(), nil
}

// ParsePage 解析分页参数，默认值是第一页10条记录
func ParsePage(pagePtr, sizePtr *int) (int, int, error) {
	page, size := 1, 10
	if pagePtr != nil {
		if *pagePtr <= 0 {
			return 0, 0, errors.New("分页页数小于1")
		}
		page = *pagePtr
	}
	if sizePtr != nil {
		if *sizePtr <= 0 {
			return 0, 0, errors.New("分页大小小于1")
		}
		size = *sizePtr
	}
	return page, size, nil
}
