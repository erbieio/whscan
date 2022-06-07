package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"server/common/types"
)

// HexToBigInt 将不带0x前缀的16进制字符串转换成大数BigInt（非法输入会返回0）
func HexToBigInt(hex string) types.BigInt {
	b := new(big.Int)
	b.SetString(hex, 16)
	return types.BigInt(b.String())
}

// HexToUint256 将不带0x前缀的16进制字符串转换为256位BigInt（大于截断后面的，非法输入会返回0）
func HexToUint256(hex string) types.BigInt {
	if len(hex) > 64 {
		hex = hex[:64]
	}
	b := new(big.Int)
	b.SetString(hex, 16)
	return types.BigInt(b.String())
}

// HexToAddress 将不带0x前缀的16进制字符串转换为Address（大于截断前面的）
func HexToAddress(hex string) types.Address {
	if len(hex) < 40 {
		hex = "0000000000000000000000000000000000000000" + hex
	}
	return types.Address("0x" + hex[len(hex)-40:])
}

// ParseAddress 将带前缀0x的16进制的字符串转换成地址
func ParseAddress(hex string) (types.Address, error) {
	if len(hex) != 42 {
		return "", fmt.Errorf("长度不是42")
	}
	if hex[0] != '0' || (hex[1] != 'x' && hex[1] != 'X') {
		return "", fmt.Errorf("前缀不是0x")
	}
	for i, c := range []byte(hex) {
		if '0' <= c && c <= '9' {
			continue
		}
		if 'a' <= c && c <= 'f' {
			continue
		}
		if 'A' <= c && c <= 'F' {
			[]byte(hex)[i] = c - 27
			continue
		}
		if 'X' == c || 'x' == c {
			[]byte(hex)[i] = 'x'
			continue
		}
		return "", fmt.Errorf("非法字符:%v", c)
	}
	return types.Address(hex), nil
}

// BigToAddress 大数转换成地址（数太大会截断前面的）
func BigToAddress(big *big.Int) types.Address {
	addr := "0000000000000000000000000000000000000000"
	if big != nil {
		addr += big.Text(16)
	}
	return types.Address("0x" + addr[len(addr)-40:])
}

// ParsePage 解析分页参数，默认值是第一页10条记录
func ParsePage(pagePtr, sizePtr *int) (int, int, error) {
	page, size := 1, 10
	if pagePtr != nil {
		page = *pagePtr
		if page <= 0 {
			return 0, 0, fmt.Errorf("分页页数小于1")
		}
	}
	if sizePtr != nil {
		size = *sizePtr
		if size <= 0 {
			return 0, 0, fmt.Errorf("分页大小小于1")
		}
		if size > 100 {
			return 0, 0, fmt.Errorf("分页大小大于100")
		}
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

func VerifyEmailFormat(email string) bool {
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}
