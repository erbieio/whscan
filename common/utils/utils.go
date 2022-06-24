package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"server/common/model"
	"server/common/types"
)

var (
	erc20TransferEventId         = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc721TransferEventId        = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc1155TransferSingleEventId = "0x7b912cc6629daab379d004780e875cdb7625e8331d3a7c8fbe08a42156325546"
	erc1155TransferBatchEventId  = "0x20114eb39ee5dfdb13684c7d9e951052ef22c89bff67131a9bf08879189b0f71"
)

// Unpack20TransferLog 解析ERC20的转移事件
func Unpack20TransferLog(log *model.Log) (*model.ERC20Transfer, error) {
	if len(log.Topics) != 3 {
		return nil, fmt.Errorf("事件主题不是2个")
	}
	if log.Topics[0] != erc20TransferEventId {
		return nil, fmt.Errorf("事件签名不匹配")
	}
	if len(log.Data) != 66 {
		return nil, fmt.Errorf("事件数据不是32字节")
	}
	return &model.ERC20Transfer{
		TxHash:  log.TxHash,
		Address: log.Address,
		From:    types.Address("0x" + log.Topics[1][26:]),
		To:      types.Address("0x" + log.Topics[2][26:]),
		Value:   HexToBigInt(log.Data[2:66]),
	}, nil
}

// Unpack721TransferLog 解析ERC721的转移事件
func Unpack721TransferLog(log *model.Log) (*model.ERC721Transfer, error) {
	if len(log.Topics) != 4 {
		return nil, fmt.Errorf("事件主题不是3个")
	}
	if log.Topics[0] != erc721TransferEventId {
		return nil, fmt.Errorf("事件签名不匹配")
	}
	if len(log.Data) != 2 {
		return nil, fmt.Errorf("事件数据不是0字节")
	}
	return &model.ERC721Transfer{
		TxHash:  log.TxHash,
		Address: log.Address,
		From:    types.Address("0x" + log.Topics[1][26:]),
		To:      types.Address("0x" + log.Topics[2][26:]),
		TokenId: HexToBigInt(log.Topics[3][2:]),
	}, nil
}

// Unpack1155TransferLog 解析ERC1155的转移（批量）事件
func Unpack1155TransferLog(log *model.Log) ([]*model.ERC1155Transfer, error) {
	if len(log.Topics) != 4 {
		return nil, fmt.Errorf("事件主题不是3个")
	}
	operator, from, to := types.Address("0x"+log.Topics[1][26:]), types.Address("0x"+log.Topics[2][26:]), types.Address("0x"+log.Topics[3][26:])

	// ERC1155 单个转移事件
	if log.Topics[0] == erc1155TransferSingleEventId {
		if len(log.Data) != 130 {
			return nil, fmt.Errorf("事件数据不是64字节")
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

	// ERC1155 批量转移事件
	if log.Topics[0] != erc1155TransferBatchEventId {
		// 动态数据类型编解码参考https://docs.soliditylang.org/en/v0.8.13/abi-spec.html#argument-encoding
		// 字长为256位即32个字节
		wordLen := (len(log.Data) - 2) / 64
		if wordLen < 4 {
			return nil, fmt.Errorf("数据少于4个字")
		}
		if wordLen%2 != 0 {
			return nil, fmt.Errorf("数据的字个数不是双数")
		}
		if log.Data[2:66] != "0000000000000000000000000000000000000000000000000000000000000040" {
			return nil, fmt.Errorf("第一个字不是0x40")
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
	return nil, fmt.Errorf("事件签名不匹配")
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
	data, err := hex.DecodeString(hexStr[130:int(130+strLen.Int64()*2)])
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
func ABIDecodeUint256(hexStr string) (Uint256, error) {
	if len(hexStr) != 66 {
		return "", fmt.Errorf("返回数据格式错误")
	}

	return Uint256(hexStr), nil
}

// ABIDecodeBool 从合约返回值只有一个的返回数据里解析bool
func ABIDecodeBool(hexStr string) (bool, error) {
	hexLen := len(hexStr)
	if hexLen != 66 || hexStr[:65] != "0x000000000000000000000000000000000000000000000000000000000000000" {
		return false, fmt.Errorf("返回数据格式错误")
	}
	return hexStr[65] == '1', nil
}

// HexToBigInt 将不带0x前缀的16进制字符串转换成大数BigInt（非法输入会返回0）
func HexToBigInt(hex string) types.BigInt {
	b := new(big.Int)
	b.SetString(hex, 16)
	return types.BigInt(b.Text(10))
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

func VerifyEmailFormat(email string) bool {
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}
