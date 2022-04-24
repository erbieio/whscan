package utils

import (
	"context"
	"fmt"
	"strings"

	"server/common/model"
	"server/common/types"
)

var (
	erc20TransferEventId         = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc721TransferEventId        = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc1155TransferSingleEventId = "0x7b912cc6629daab379d004780e875cdb7625e8331d3a7c8fbe08a42156325546"
	erc1155TransferBatchEventId  = "0x20114eb39ee5dfdb13684c7d9e951052ef22c89bff67131a9bf08879189b0f71"

	notInterfaceId     = types.Bytes4("0xffffffff")
	erc165InterfaceId  = types.Bytes4("0x01ffc9a7")
	erc721InterfaceId  = types.Bytes4("0x80ac58cd")
	erc1155InterfaceId = types.Bytes4("0xd9b67a26")
	Addr1              = types.Address("0x0000000000000000000000000000000000000001")
	Addr2              = types.Address("0x0000000000000000000000000000000000000002")
)

type ContractClient interface {
	CallContract(ctx context.Context, msg map[string]interface{}, number *types.BigInt) (types.Bytes, error)
}

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
		Value:   HexToUint256(log.Data[2:66]),
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
		TokenId: HexToUint256(log.Topics[3][2:]),
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
			TokenId:  HexToUint256(log.Data[2:66]),
			Value:    HexToUint256(log.Data[66:130]),
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
				TokenId:  HexToUint256(log.Data[idOffset : idOffset+64]),
				Value:    HexToUint256(log.Data[valueOffset : valueOffset+64]),
			}
		}
		return transferLogs, nil
	}
	return nil, fmt.Errorf("事件签名不匹配")
}

func IsERC165(client ContractClient, address types.Address) (bool, error) {
	support, err := SupportsInterface(client, address, erc165InterfaceId)
	if !support || err != nil {
		return false, FilterContractErr(err)
	}
	support, err = SupportsInterface(client, address, notInterfaceId)
	return !support, FilterContractErr(err)
}

func IsERC721(client ContractClient, address types.Address) (bool, error) {
	support, err := SupportsInterface(client, address, erc721InterfaceId)
	return support, FilterContractErr(err)
}

func IsERC1155(client ContractClient, address types.Address) (bool, error) {
	support, err := SupportsInterface(client, address, erc1155InterfaceId)
	return support, FilterContractErr(err)
}

func IsERC20(client ContractClient, address types.Address) (bool, error) {
	_, err := TotalSupply(client, address)
	if err != nil {
		return false, FilterContractErr(err)
	}
	_, err = BalanceOf(client, address, Addr1)
	if err != nil {
		return false, FilterContractErr(err)
	}
	_, err = Allowance(client, address, Addr1, Addr2)
	if err != nil {
		return false, FilterContractErr(err)
	}
	_, err = Transfer(client, address, Addr2, types.Big0, &Addr1)
	if err != nil {
		return false, FilterContractErr(err)
	}
	_, err = TransferFrom(client, address, Addr1, Addr2, types.Big0, &Addr1)
	if err != nil {
		return false, FilterContractErr(err)
	}
	ok, err := Approve(client, address, Addr2, types.Big0, &Addr1)
	if !ok || err != nil {
		return false, FilterContractErr(err)
	}
	return true, nil
}

// GetERC 获取合约类型，合约调用的错误将不会返回且合约将视为无类型的合约
func GetERC(client ContractClient, address types.Address) (types.ERC, error) {
	ok, err := IsERC165(client, address)
	if err != nil {
		return types.NONE, err
	}
	if !ok {
		ok, err = IsERC20(client, address)
		if ok && err == nil {
			return types.ERC20, nil
		} else {
			return types.NONE, err
		}
	}
	ok, err = IsERC721(client, address)
	if err != nil {
		return types.NONE, err
	}
	if ok {
		return types.ERC721, nil
	}
	ok, err = IsERC1155(client, address)
	if err != nil {
		return types.NONE, err
	}
	if ok {
		return types.ERC1155, nil
	}
	return types.ERC165, nil
}

// FilterContractErr 过滤掉除网络连接外的错误
func FilterContractErr(err error) error {
	if err != nil {
		if strings.Index(err.Error(), "connection") > 0 {
			return err
		}
		if strings.Index(err.Error(), "unexpected EOF") > 0 {
			return err
		}
	}
	return nil
}
