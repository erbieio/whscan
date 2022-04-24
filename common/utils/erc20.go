package utils

import (
	"context"
	"server/common/types"
)

var (
	nameSelector         = "0x06fdde03"
	symbolSelector       = "0x95d89b41"
	decimalsSelector     = "0x313ce567"
	totalSupplySelector  = "0x18160ddd"
	allowanceSelector    = "0xdd62ed3e"
	balanceOfSelector    = "0x70a08231"
	transferSelector     = "0xa9059cbb"
	transferFromSelector = "0x23b872dd"
	approveSelector      = "0x095ea7b3"
)

// Name 查询给定ERC20合约的代币名称（可选接口）
func Name(client ContractClient, address types.Address) (string, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": nameSelector,
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return *new(string), err
	}

	return ABIDecodeString(string(out))
}

// Symbol 查询给定ERC20合约的代币符号（可选接口）
func Symbol(client ContractClient, address types.Address) (string, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": symbolSelector,
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return *new(string), err
	}

	return ABIDecodeString(string(out))
}

// Decimals 查询给定ERC20合约的代币符号（可选接口）
func Decimals(client ContractClient, address types.Address) (uint8, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": decimalsSelector,
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, err
	}

	return ABIDecodeUint8(string(out))
}

// TotalSupply 查询给定ERC20合约的代币发行总量（必须接口）
func TotalSupply(client ContractClient, address types.Address) (types.Uint256, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": totalSupplySelector,
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return "", err
	}

	return ABIDecodeUint256(string(out))
}

func Allowance(client ContractClient, address, owner, spender types.Address) (types.Uint256, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": allowanceSelector + "000000000000000000000000" + string(owner[2:]) + "000000000000000000000000" + string(spender[2:]),
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return "", err
	}

	return ABIDecodeUint256(string(out))
}

func BalanceOf(client ContractClient, address, account types.Address) (types.Uint256, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": balanceOfSelector + "000000000000000000000000" + string(account[2:]),
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return "", err
	}

	return ABIDecodeUint256(string(out))
}

func Transfer(client ContractClient, address, to types.Address, amount types.Uint256, caller *types.Address) (bool, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": transferSelector + "000000000000000000000000" + string(to[2:]) + string(amount[2:]),
	}
	if caller != nil {
		msg["from"] = *caller
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, err
	}

	return ABIDecodeBool(string(out))
}

func TransferFrom(client ContractClient, address, from, to types.Address, amount types.Uint256, caller *types.Address) (bool, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": transferFromSelector + "000000000000000000000000" + string(from[2:]) + "000000000000000000000000" + string(to[2:]) + string(amount[2:]),
	}
	if caller != nil {
		msg["from"] = *caller
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, err
	}

	return ABIDecodeBool(string(out))
}

func Approve(client ContractClient, address, spender types.Address, amount types.Uint256, caller *types.Address) (bool, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": approveSelector + "000000000000000000000000" + string(spender[2:]) + string(amount[2:]),
	}
	if caller != nil {
		msg["from"] = *caller
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, err
	}

	return ABIDecodeBool(string(out))
}
