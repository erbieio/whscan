package utils

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"

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

	//erc721
	tokenURISelector = "0xc87b56dd"
	ownerOfSelector  = "0x6352211e"

	//erc1155
	balanceOf1155Selector = "0x00fdd58e"
)

type ContractClient interface {
	CallContract(ctx context.Context, to, data, number any) (types.Bytes, error)
}

// SupportsInterface query whether a given contract supports interfaceId
func SupportsInterface(c ContractClient, ctx context.Context, number, address any, interfaceId string) (bool, error) {
	data := supportsInterfaceSelector + interfaceId[2:10] + "00000000000000000000000000000000000000000000000000000000"
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return false, err
	}
	return ABIDecodeBool(out)
}

// Name Query the token name of a given ERC20 contract (optional interface)
func Name(c ContractClient, ctx context.Context, number, address any) (string, error) {
	out, err := c.CallContract(ctx, address, nameSelector, number)
	if err != nil {
		return "", err
	}
	return ABIDecodeString(out)
}

// Symbol query the token symbol of a given ERC20 contract (optional interface)
func Symbol(c ContractClient, ctx context.Context, number, address any) (string, error) {
	out, err := c.CallContract(ctx, address, symbolSelector, number)
	if err != nil {
		return "", err
	}
	return ABIDecodeString(out)
}

// Decimals query the token symbol of a given ERC20 contract (optional interface)
func Decimals(c ContractClient, ctx context.Context, number, address any) (uint8, error) {
	out, err := c.CallContract(ctx, address, decimalsSelector, number)
	if err != nil {
		return 0, err
	}
	return ABIDecodeUint8(out)
}

// TotalSupply queries the total amount of tokens issued for a given ERC20 contract (required interface)
func TotalSupply(c ContractClient, ctx context.Context, number, address any) (string, error) {
	out, err := c.CallContract(ctx, address, totalSupplySelector, number)
	if err != nil {
		return "", err
	}
	return ABIDecodeBigInt(out)
}

func Allowance(c ContractClient, ctx context.Context, number, address any, owner, spender string) (string, error) {
	data := allowanceSelector + "000000000000000000000000" + owner[2:] + "000000000000000000000000" + spender[2:]
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return "", err
	}
	return ABIDecodeBigInt(out)
}

func BalanceOf(c ContractClient, ctx context.Context, number, address any, account string) (string, error) {
	data := balanceOfSelector + "000000000000000000000000" + account[2:]
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return "", err
	}
	return ABIDecodeBigInt(out)
}

func Transfer(c ContractClient, ctx context.Context, number, address any, to, amount string) (bool, error) {
	data := transferSelector + "000000000000000000000000" + to[2:] + amount[2:]
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return false, err
	}
	return ABIDecodeBool(out)
}

func TransferFrom(c ContractClient, ctx context.Context, number, address any, from, to, amount string) (bool, error) {
	data := transferFromSelector + "000000000000000000000000" + from[2:] + "000000000000000000000000" + to[2:] + amount[2:]
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return false, err
	}
	return ABIDecodeBool(out)
}

func Approve(c ContractClient, ctx context.Context, number, address any, spender, amount string) (bool, error) {
	data := approveSelector + "000000000000000000000000" + spender[2:] + amount[2:]
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return false, err
	}
	return ABIDecodeBool(out)
}

// ABIDecodeString parses the string from the returned data with only one contract return value
func ABIDecodeString(out types.Bytes) (string, error) {
	outLen := len(out)
	if outLen < 130 || (outLen-2)%64 != 0 || out[64:66] != "20" {
		return "", errors.New("return data format error")
	}
	strLen := new(big.Int)
	strLen.SetString(string(out[66:130]), 16)
	//fmt.Println("(outLen-130)/64 = ", (outLen-130)/64, "int(strLen.Int64())/32 = ", int(strLen.Int64())/32, "strLen.Int64() = ", strLen.Int64())
	//if (outLen-130)/64 != int(strLen.Int64())/32 {
	//	return "", errors.New("return data string length error")
	//}
	data, err := hex.DecodeString(string(out[130:int(130+strLen.Int64()*2)]))
	return string(data), err
}

// ABIDecodeUint8 parses uint8 from the return data with only one contract return value
func ABIDecodeUint8(out types.Bytes) (uint8, error) {
	outLen := len(out)
	if outLen != 66 || out[:50] != "0x000000000000000000000000000000000000000000000000" {
		return 0, errors.New("return data format error")
	}
	data, err := strconv.ParseUint(string(out[50:]), 16, 8)
	return uint8(data), err
}

// ABIDecodeBigInt parses uint256 from the return data with only one contract return value
func ABIDecodeBigInt(out types.Bytes) (string, error) {
	if len(out) != 66 {
		return "", errors.New("return data format error")
	}
	return string(out), nil
}

// ABIDecodeBool parses bool from the return data with only one contract return value
func ABIDecodeBool(out types.Bytes) (bool, error) {
	outLen := len(out)
	if outLen != 66 || out[:65] != "0x000000000000000000000000000000000000000000000000000000000000000" {
		return false, errors.New("return data format error")
	}
	return out[65] == '1', nil
}

// ABIDecodeAddress parses address from the return data with only one contract return value
func ABIDecodeAddress(out types.Bytes) (string, error) {
	outLen := len(out)
	if outLen != 66 {
		return "", errors.New("return data format error")
	}
	return "0x" + string(out[26:]), nil
}

// nft

// GetTokenURI get tokenURI of a given ERC721 contract (optional interface)
func GetTokenURI(c ContractClient, ctx context.Context, number, address any, tokenId int64) (string, error) {
	strTokenId := big.NewInt(tokenId).Text(16)
	var str0 string
	for i := 0; i < 64-len(strTokenId); i++ {
		str0 = str0 + "0"
	}
	strTokenId = str0 + strTokenId

	data := tokenURISelector + strTokenId
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return "", err
	}
	return ABIDecodeString(out)
}

// GetOwnerOf get owner of a given ERC721 contract (optional interface)
func GetOwnerOf(c ContractClient, ctx context.Context, number, address any, tokenId int64) (string, error) {
	strTokenId := big.NewInt(tokenId).Text(16)
	var str0 string
	for i := 0; i < 64-len(strTokenId); i++ {
		str0 = str0 + "0"
	}
	strTokenId = str0 + strTokenId

	data := ownerOfSelector + strTokenId
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return "", err
	}
	return ABIDecodeAddress(out)
}

func BalanceOf1155(c ContractClient, ctx context.Context, number, address any, account string, tokenId int64) (string, error) {
	strTokenId := big.NewInt(tokenId).Text(16)
	var str0 string
	for i := 0; i < 64-len(strTokenId); i++ {
		str0 = str0 + "0"
	}
	strTokenId = str0 + strTokenId

	data := balanceOf1155Selector + "000000000000000000000000" + account[2:] + strTokenId
	out, err := c.CallContract(ctx, address, data, number)
	if err != nil {
		return "", err
	}
	return ABIDecodeBigInt(out)
}
