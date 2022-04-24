package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/sha3"
	"server/common/types"
)

func HexToBigInt(hex string) types.BigInt {
	b := new(big.Int)
	b.UnmarshalText([]byte(hex))
	return types.BigInt(b.String())
}

// HexToUint256 将不带0x前缀的16进制字符串转换为uint256的大数
func HexToUint256(hex string) types.BigInt {
	if len(hex) > 64 {
		hex = hex[:64]
	}
	b := new(big.Int)
	b.SetString(hex, 16)
	return types.BigInt(b.String())
}

func HexToAddress(hex string) (types.Address, error) {
	if len(hex) != 42 {
		return "", fmt.Errorf("长度不是42")
	}
	if hex[0] != '0' || (hex[1] != 'x' && hex[1] != 'X') {
		return "", fmt.Errorf("前缀不是0x")
	}
	for _, c := range hex[2:] {
		if ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F') {
			continue
		}
		return "", fmt.Errorf("非法字符:%v", c)
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

func PubkeyToAddress(p ecdsa.PublicKey) types.Address {
	data := elliptic.Marshal(secp256k1.S256(), p.X, p.Y)
	return types.Address("0x" + Keccak256(data[1:])[26:])
}

func Keccak256Hash(data ...[]byte) (h types.Hash) {
	return types.Hash(Keccak256(data...))
}

func Keccak256(data ...[]byte) (h string) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}

	return "0x" + hex.EncodeToString(d.Sum(nil))
}

func HexToECDSA(key string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(key)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, fmt.Errorf("invalid hex data for private key")
	}
	return secp256k1.PrivKeyFromBytes(b).ToECDSA(), nil
}

// ParsePage 解析分页参数，默认值是第一页10条记录
func ParsePage(pagePtr, sizePtr *int) (int, int, error) {
	page, size := 1, 10
	if pagePtr != nil {
		if *pagePtr <= 0 {
			return 0, 0, fmt.Errorf("分页页数小于1")
		}
		page = *pagePtr
	}
	if sizePtr != nil {
		if *sizePtr <= 0 {
			return 0, 0, fmt.Errorf("分页大小小于1")
		}
		size = *sizePtr
	}
	return page, size, nil
}

// ABIDecodeString 从合约返回值只有一个的返回数据里解析字符串
func ABIDecodeString(hexStr string) (string, error) {
	hexLen := len(hexStr)
	if hexLen < 130 || (hexLen-2)%64 != 0 || hexStr[64:66] != "20" {
		return "", fmt.Errorf("返回数据格式错误")
	}
	strLen := new(big.Int)
	strLen.SetString(hexStr[66:130], 16)
	if (hexLen-130)/64 != int(strLen.Int64())/32 {
		return "", fmt.Errorf("返回数据字符串长度错误")
	}
	data, err := hex.DecodeString(hexStr[130:int(strLen.Int64()*2)])
	return string(data), err
}

// ABIDecodeUint8 从合约返回值只有一个的返回数据里解析uint8
func ABIDecodeUint8(hexStr string) (uint8, error) {
	hexLen := len(hexStr)
	if hexLen != 66 || hexStr[:50] != "0x000000000000000000000000000000000000000000000000" {
		return 0, fmt.Errorf("返回数据格式错误")
	}
	data, err := strconv.ParseUint(hexStr[50:], 16, 8)
	return uint8(data), err
}

// ABIDecodeUint256 从合约返回值只有一个的返回数据里解析uint256
func ABIDecodeUint256(hexStr string) (types.Uint256, error) {
	if len(hexStr) != 66 {
		return "", fmt.Errorf("返回数据格式错误")
	}

	return types.Uint256(hexStr), nil
}

// ABIDecodeBool 从合约返回值只有一个的返回数据里解析bool
func ABIDecodeBool(hexStr string) (bool, error) {
	hexLen := len(hexStr)
	if hexLen != 66 || hexStr[:65] != "0x000000000000000000000000000000000000000000000000000000000000000" {
		return false, fmt.Errorf("返回数据格式错误")
	}
	return hexStr[65] == '1', nil
}
