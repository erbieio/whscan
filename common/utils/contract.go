package utils

import (
	"context"
	"strings"

	"server/common/types"
)

var (
	supportsInterfaceSelector = "0x01ffc9a7"
	nameSelector              = "0x06fdde03"
	symbolSelector            = "0x95d89b41"
	decimalsSelector          = "0x313ce567"
	totalSupplySelector       = "0x18160ddd"
	allowanceSelector         = "0xdd62ed3e"
	balanceOfSelector         = "0x70a08231"
	transferSelector          = "0xa9059cbb"
	transferFromSelector      = "0x23b872dd"
	approveSelector           = "0x095ea7b3"

	notInterfaceId     = "0xffffffff"
	erc165InterfaceId  = "0x01ffc9a7"
	erc721InterfaceId  = "0x80ac58cd"
	erc1155InterfaceId = "0xd9b67a26"

	Addr1 = types.Address("0x0000000000000000000000000000000000000001")
	Addr2 = types.Address("0x0000000000000000000000000000000000000002")
	Big0  = Uint256("0x0000000000000000000000000000000000000000000000000000000000000000")
)

// Uint256 带前缀0x和前导0的16进制字符串uint256
type Uint256 string

type ContractClient interface {
	CallContract(ctx context.Context, msg map[string]interface{}, number *types.BigInt) (types.Data, error)
}

// SupportsInterface 查询给定合约是否支持interfaceId
func SupportsInterface(client ContractClient, address types.Address, interfaceId string) (bool, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": supportsInterfaceSelector + interfaceId[2:10] + "00000000000000000000000000000000000000000000000000000000",
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, err
	}

	return ABIDecodeBool(string(out))
}

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
func TotalSupply(client ContractClient, address types.Address) (Uint256, error) {
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

func Allowance(client ContractClient, address, owner, spender types.Address) (Uint256, error) {
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

func BalanceOf(client ContractClient, address, account types.Address) (Uint256, error) {
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

func Transfer(client ContractClient, address, to types.Address, amount Uint256, caller *types.Address) (bool, error) {
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

func TransferFrom(client ContractClient, address, from, to types.Address, amount Uint256, caller *types.Address) (bool, error) {
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

func Approve(client ContractClient, address, spender types.Address, amount Uint256, caller *types.Address) (bool, error) {
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

func Property(client ContractClient, address types.Address) (name, symbol *string, err error) {
	n, err := Name(client, address)
	if err == nil {
		name = &n
	}
	err = filterContractErr(err)
	if err != nil {
		return
	}
	s, err := Symbol(client, address)
	if err == nil {
		symbol = &s
	}
	err = filterContractErr(err)
	return
}

func IsERC165(client ContractClient, address types.Address) (bool, error) {
	support, err := SupportsInterface(client, address, erc165InterfaceId)
	if !support || err != nil {
		return false, filterContractErr(err)
	}
	support, err = SupportsInterface(client, address, notInterfaceId)
	return !support, filterContractErr(err)
}

func IsERC721(client ContractClient, address types.Address) (bool, error) {
	support, err := SupportsInterface(client, address, erc721InterfaceId)
	return support, filterContractErr(err)
}

func IsERC1155(client ContractClient, address types.Address) (bool, error) {
	support, err := SupportsInterface(client, address, erc1155InterfaceId)
	return support, filterContractErr(err)
}

func IsERC20(client ContractClient, address types.Address) (bool, error) {
	_, err := TotalSupply(client, address)
	if err != nil {
		return false, filterContractErr(err)
	}
	_, err = BalanceOf(client, address, Addr1)
	if err != nil {
		return false, filterContractErr(err)
	}
	_, err = Allowance(client, address, Addr1, Addr2)
	if err != nil {
		return false, filterContractErr(err)
	}
	_, err = Transfer(client, address, Addr2, Big0, &Addr1)
	if err != nil {
		return false, filterContractErr(err)
	}
	_, err = TransferFrom(client, address, Addr1, Addr2, Big0, &Addr1)
	if err != nil {
		return false, filterContractErr(err)
	}
	ok, err := Approve(client, address, Addr2, Big0, &Addr1)
	if !ok || err != nil {
		return false, filterContractErr(err)
	}
	return true, nil
}

// filterContractErr 过滤掉除网络连接外的错误
func filterContractErr(err error) error {
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
