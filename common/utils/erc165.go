package utils

import (
	"context"
	"server/common/types"
)

var supportsInterfaceSelector = "0x01ffc9a7"

// SupportsInterface 查询给定合约是否支持interfaceId
func SupportsInterface(client ContractClient, address types.Address, interfaceId types.Bytes4) (bool, error) {
	msg := map[string]interface{}{
		"to":   address,
		"data": supportsInterfaceSelector + string(interfaceId[2:10]) + "00000000000000000000000000000000000000000000000000000000",
	}
	out, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, err
	}

	return ABIDecodeBool(string(out))
}
