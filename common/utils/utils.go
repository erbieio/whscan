package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"time"

	"server/common/model"
	"server/common/types"
)

var (
	erc20TransferEventId         = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc721TransferEventId        = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc1155TransferSingleEventId = "0x7b912cc6629daab379d004780e875cdb7625e8331d3a7c8fbe08a42156325546"
	erc1155TransferBatchEventId  = "0x20114eb39ee5dfdb13684c7d9e951052ef22c89bff67131a9bf08879189b0f71"
	loc, _                       = time.LoadLocation("Local")
	DaySecond                    = 24 * time.Hour.Milliseconds() / 1000
)

// Unpack20TransferLog parses ERC20 transfer events
func Unpack20TransferLog(log *model.Log) (*model.ERC20Transfer, error) {
	if len(log.Topics) != 3 {
		return nil, fmt.Errorf("The event subject is not 2")
	}
	if log.Topics[0] != erc20TransferEventId {
		return nil, fmt.Errorf("Event signature does not match")
	}
	if len(log.Data) != 66 {
		return nil, fmt.Errorf("Event data is not 32 bytes")
	}
	return &model.ERC20Transfer{
		TxHash:  log.TxHash,
		Address: log.Address,
		From:    types.Address("0x" + log.Topics[1][26:]),
		To:      types.Address("0x" + log.Topics[2][26:]),
		Value:   HexToBigInt(log.Data[2:66]),
	}, nil
}

// Unpack721TransferLog parses ERC721 transfer events
func Unpack721TransferLog(log *model.Log) (*model.ERC721Transfer, error) {
	if len(log.Topics) != 4 {
		return nil, fmt.Errorf("The event subject is not 3")
	}
	if log.Topics[0] != erc721TransferEventId {
		return nil, fmt.Errorf("Event signature does not match")
	}
	if len(log.Data) != 2 {
		return nil, fmt.Errorf("Event data is not 0 bytes")
	}
	return &model.ERC721Transfer{
		TxHash:  log.TxHash,
		Address: log.Address,
		From:    types.Address("0x" + log.Topics[1][26:]),
		To:      types.Address("0x" + log.Topics[2][26:]),
		TokenId: HexToBigInt(log.Topics[3][2:]),
	}, nil
}

// Unpack1155TransferLog parses ERC1155 transfer (batch) events
func Unpack1155TransferLog(log *model.Log) ([]*model.ERC1155Transfer, error) {
	if len(log.Topics) != 4 {
		return nil, fmt.Errorf("The event subject is not 3")
	}
	operator, from, to := types.Address("0x"+log.Topics[1][26:]), types.Address("0x"+log.Topics[2][26:]), types.Address("0x"+log.Topics[3][26:])

	// ERC1155 single transfer event
	if log.Topics[0] == erc1155TransferSingleEventId {
		if len(log.Data) != 130 {
			return nil, fmt.Errorf("Event data is not 64 bytes")
		}
		return []*model.ERC1155Transfer{{
			TxHash:   log.TxHash,
			Address:  log.Address,
			Operator: operator,
			From:     from,
			To:       to,
			TokenId:  HexToBigInt(log.Data[2:66]),
			Value:    HexToBigInt(log.Data[66:130]),
		}}, nil
	}

	// ERC1155 batch transfer event
	if log.Topics[0] != erc1155TransferBatchEventId {
		// Dynamic data type encoding and decoding reference https://docs.soliditylang.org/en/v0.8.13/abi-spec.html#argument-encoding
		// The word length is 256 bits or 32 bytes
		wordLen := (len(log.Data) - 2) / 64
		if wordLen < 4 {
			return nil, fmt.Errorf("The data is less than 4 words")
		}
		if wordLen%2 != 0 {
			return nil, fmt.Errorf("The number of words in the data is not a double number")
		}
		if log.Data[2:66] != "0000000000000000000000000000000000000000000000000000000000000040" {
			return nil, fmt.Errorf("The first word is not 0x40")
		}
		transferCount := (wordLen - 4) / 2
		transferLogs := make([]*model.ERC1155Transfer, transferCount)
		for i := 0; i < transferCount; i++ {
			idOffset, valueOffset := 2+(i+3)*64, 2+(transferCount+i+4)*64
			transferLogs[i] = &model.ERC1155Transfer{
				TxHash:   log.TxHash,
				Address:  log.Address,
				Operator: operator,
				From:     from,
				To:       to,
				TokenId:  HexToBigInt(log.Data[idOffset : idOffset+64]),
				Value:    HexToBigInt(log.Data[valueOffset : valueOffset+64]),
			}
		}
		return transferLogs, nil
	}
	return nil, fmt.Errorf("Event signature does not match")
}

// ABIDecodeString parses the string from the return data with only one return value from the contract
func ABIDecodeString(hexStr string) (string, error) {
	hexLen := len(hexStr)
	if hexLen < 130 || (hexLen-2)%64 != 0 || hexStr[64:66] != "20" {
		return "", fmt.Errorf("Return data format error")
	}
	strLen := new(big.Int)
	strLen.SetString(hexStr[66:130], 16)
	if (hexLen-130)/64 != int(strLen.Int64())/32 {
		return "", fmt.Errorf("Return data string length error")
	}
	data, err := hex.DecodeString(hexStr[130:int(130+strLen.Int64()*2)])
	return string(data), err
}

// ABIDecodeUint8 parses uint8 from the return data with only one return value from the contract
func ABIDecodeUint8(hexStr string) (uint8, error) {
	hexLen := len(hexStr)
	if hexLen != 66 || hexStr[:50] != "0x0000000000000000000000000000000000000000000000000" {
		return 0, fmt.Errorf("Return data format error")
	}
	data, err := strconv.ParseUint(hexStr[50:], 16, 8)
	return uint8(data), err
}

// ABIDecodeUint256 parses uint256 from the return data with only one return value from the contract
func ABIDecodeUint256(hexStr string) (Uint256, error) {
	if len(hexStr) != 66 {
		return "", fmt.Errorf("Return data format error")
	}

	return Uint256(hexStr), nil
}

// ABIDecodeBool parses the bool from the return data with only one return value from the contract
func ABIDecodeBool(hexStr string) (bool, error) {
	hexLen := len(hexStr)
	if hexLen != 66 || hexStr[:65] != "0x000000000000000000000000000000000000000000000000000000000000000" {
		return false, fmt.Errorf("Return data format error")
	}
	return hexStr[65] == '1', nil
}

// HexToBigInt converts a hexadecimal string without 0x prefix to a big number BigInt (illegal input will return 0)
func HexToBigInt(hex string) types.BigInt {
	b := new(big.Int)
	b.SetString(hex, 16)
	return types.BigInt(b.Text(10))
}

// HexToAddress converts a hexadecimal string without a 0x prefix to an Address (greater than the truncated front)
func HexToAddress(hex string) types.Address {
	if len(hex) < 40 {
		hex = "0000000000000000000000000000000000000000" + hex
	}
	return types.Address("0x" + hex[len(hex)-40:])
}

// ParseAddress converts a hexadecimal string prefixed with 0x to an address
func ParseAddress(hex string) (types.Address, error) {
	if len(hex) != 42 {
		return "", fmt.Errorf("Length is not 42")
	}
	if hex[0] != '0' || (hex[1] != 'x' && hex[1] != 'X') {
		return "", fmt.Errorf("Prefix is ​​not 0x")
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
		return "", fmt.Errorf("Illegal character: %v", c)
	}
	return types.Address(hex), nil
}

// BigToAddress converts large numbers into addresses (too large numbers will truncate the previous ones)
func BigToAddress(big *big.Int) types.Address {
	addr := "0000000000000000000000000000000000000000"
	if big != nil {
		addr += big.Text(16)
	}
	return types.Address("0x" + addr[len(addr)-40:])
}

// ParsePage parses the paging parameters, the default value is 10 records on the first page
func ParsePage(pagePtr, sizePtr *int) (int, int, error) {
	page, size := 1, 10
	if pagePtr != nil {
		page = *pagePtr
		if page <= 0 {
			return 0, 0, fmt.Errorf("Number of paging pages is less than 1")
		}
	}
	if sizePtr != nil {
		size = *sizePtr
		if size <= 0 {
			return 0, 0, fmt.Errorf("The page size is less than 1")
		}
		if size > 100 {
			return 0, 0, fmt.Errorf("Page size is greater than 100")
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
