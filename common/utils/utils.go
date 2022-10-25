package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "server/common/model"
	. "server/common/types"
)

var (
	erc20TransferEventId         = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc721TransferEventId        = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc1155TransferSingleEventId = "0x7b912cc6629daab379d004780e875cdb7625e8331d3a7c8fbe08a42156325546"
	erc1155TransferBatchEventId  = "0x20114eb39ee5dfdb13684c7d9e951052ef22c89bff67131a9bf08879189b0f71"
	loc, _                       = time.LoadLocation("Local")
	DaySecond                    = 24 * time.Hour.Milliseconds() / 1000
)

func UnpackTransferLog(log *Log) []interface{} {
	topicsLen := len(log.Topics)
	if topicsLen == 3 {
		if log.Topics[0] == erc20TransferEventId && len(log.Data) == 66 {
			// 解析ERC20的转移事件
			return []interface{}{&ERC20Transfer{
				TxHash:  log.TxHash,
				Address: log.Address,
				From:    Address("0x" + log.Topics[1][26:]),
				To:      Address("0x" + log.Topics[2][26:]),
				Value:   HexToBigInt(log.Data[2:66]),
			}}
		}
	} else if topicsLen == 4 {
		if log.Topics[0] == erc721TransferEventId && len(log.Data) == 2 {
			// 解析ERC721的转移事件
			return []interface{}{&ERC721Transfer{
				TxHash:  log.TxHash,
				Address: log.Address,
				From:    Address("0x" + log.Topics[1][26:]),
				To:      Address("0x" + log.Topics[2][26:]),
				TokenId: HexToBigInt(log.Topics[3][2:]),
			}}
		} else if log.Topics[0] == erc1155TransferSingleEventId && len(log.Data) == 130 {
			// 解析ERC1155的单个转移事件
			operator, from, to := Address("0x"+log.Topics[1][26:]), Address("0x"+log.Topics[2][26:]), Address("0x"+log.Topics[3][26:])
			return []interface{}{&ERC1155Transfer{
				TxHash:   log.TxHash,
				Address:  log.Address,
				Operator: operator,
				From:     from,
				To:       to,
				TokenId:  HexToBigInt(log.Data[2:66]),
				Value:    HexToBigInt(log.Data[66:130]),
			}}
		} else if log.Topics[0] == erc1155TransferBatchEventId {
			// 解析ERC1155的批量转移事件
			operator, from, to := Address("0x"+log.Topics[1][26:]), Address("0x"+log.Topics[2][26:]), Address("0x"+log.Topics[3][26:])
			// 动态数据类型编解码参考https://docs.soliditylang.org/en/v0.8.13/abi-spec.html#argument-encoding
			// 字长为256位即32个字节
			wordLen := (len(log.Data) - 2) / 64
			if wordLen < 4 || wordLen%2 != 0 || log.Data[2:66] != "0000000000000000000000000000000000000000000000000000000000000040" {
				return nil
			}
			transferCount := (wordLen - 4) / 2
			transferLogs := make([]interface{}, transferCount)
			for i := 0; i < transferCount; i++ {
				idOffset, valueOffset := 2+(i+3)*64, 2+(transferCount+i+4)*64
				transferLogs[i] = &ERC1155Transfer{
					TxHash:   log.TxHash,
					Address:  log.Address,
					Operator: operator,
					From:     from,
					To:       to,
					TokenId:  HexToBigInt(log.Data[idOffset : idOffset+64]),
					Value:    HexToBigInt(log.Data[valueOffset : valueOffset+64]),
				}
			}
			return transferLogs
		}
	}
	return nil
}

// ABIDecodeString parses the string from the return data with only one return value from the contract
func ABIDecodeString(hexStr string) (string, error) {
	hexLen := len(hexStr)
	if hexLen < 130 || (hexLen-2)%64 != 0 || hexStr[64:66] != "20" {
		return "", fmt.Errorf("return data format error")
	}
	strLen := new(big.Int)
	strLen.SetString(hexStr[66:130], 16)
	if (hexLen-130)/64 != int(strLen.Int64())/32 {
		return "", fmt.Errorf("return data string length error")
	}
	data, err := hex.DecodeString(hexStr[130:int(130+strLen.Int64()*2)])
	return string(data), err
}

// ABIDecodeUint8 parses uint8 from the return data with only one return value from the contract
func ABIDecodeUint8(hexStr string) (uint8, error) {
	hexLen := len(hexStr)
	if hexLen != 66 || hexStr[:50] != "0x0000000000000000000000000000000000000000000000000" {
		return 0, fmt.Errorf("return data format error")
	}
	data, err := strconv.ParseUint(hexStr[50:], 16, 8)
	return uint8(data), err
}

// ABIDecodeUint256 parses uint256 from the return data with only one return value from the contract
func ABIDecodeUint256(hexStr string) (Uint256, error) {
	if len(hexStr) != 66 {
		return "", fmt.Errorf("return data format error")
	}

	return Uint256(hexStr), nil
}

// ABIDecodeBool parses the bool from the return data with only one return value from the contract
func ABIDecodeBool(hexStr string) (bool, error) {
	hexLen := len(hexStr)
	if hexLen != 66 || hexStr[:65] != "0x000000000000000000000000000000000000000000000000000000000000000" {
		return false, fmt.Errorf("return data format error")
	}
	return hexStr[65] == '1', nil
}

// HexToBigInt converts a hexadecimal string without 0x prefix to a big number BigInt (illegal input will return 0)
func HexToBigInt(hex string) BigInt {
	b := new(big.Int)
	b.SetString(hex, 16)
	return BigInt(b.Text(10))
}

// HexToAddress converts a hexadecimal string without a 0x prefix to an Address (greater than the truncated front)
func HexToAddress(hex string) *Address {
	if len(hex) < 42 {
		hex = "0x000000000000000000000000000000000000000" + hex[2:]
	}
	if len(hex) > 42 {
		hex = "0x" + hex[len(hex)-40:]
	}
	return (*Address)(&hex)
}

// ParseAddress converts a hexadecimal string prefixed with 0x to an address
func ParseAddress(hex []byte) (Address, error) {
	if len(hex) != 42 {
		return "", fmt.Errorf("length is not 42")
	}
	if hex[0] != '0' || (hex[1] != 'x' && hex[1] != 'X') {
		return "", fmt.Errorf("prefix is not 0x")
	}
	hex[1] = 'x'
	for i := 2; i < 42; i++ {
		c := hex[i]
		if ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') {
			continue
		}
		if 'A' <= c && c <= 'F' {
			hex[i] = c + 32
			continue
		}
		return "", fmt.Errorf("illegal character: %v", c)
	}
	return Address(hex), nil
}

// BigToAddress converts large numbers into addresses (too large numbers will truncate the previous ones)
func BigToAddress(big *big.Int) Address {
	addr := "0000000000000000000000000000000000000000"
	if big != nil {
		addr += big.Text(16)
	}
	return Address("0x" + addr[len(addr)-40:])
}

func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := HomeDir(); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// ParsePage parses the paging parameters, the default value is 10 records on the first page
func ParsePage(pagePtr, sizePtr *int) (int, int, error) {
	page, size := 1, 10
	if pagePtr != nil {
		page = *pagePtr
		if page <= 0 {
			return 0, 0, fmt.Errorf("number of paging pages is less than 1")
		}
	}
	if sizePtr != nil {
		size = *sizePtr
		if size <= 0 {
			return 0, 0, fmt.Errorf("the page size is less than 1")
		}
		if size > 100 {
			return 0, 0, fmt.Errorf("page size is greater than 100")
		}
	}
	return page, size, nil
}

// LastTimeRange unix time range for the specified number of days based on the current time
func LastTimeRange(day int64) (start, stop int64) {
	now := time.Now().Local()
	stopTime, _ := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02")+" 00:00:00", loc)
	stop = stopTime.Unix()
	start = stop - DaySecond*day
	return
}

func VerifyEmailFormat(email string) bool {
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}
