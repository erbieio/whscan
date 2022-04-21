package utils

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ERC165ABI          abi.ABI
	NotInterfaceId     = [4]byte{0xff, 0xff, 0xff, 0xff}
	ERC165InterfaceId  = [4]byte{0x01, 0xff, 0xc9, 0xa7}
	ERC721InterfaceId  = [4]byte{0x80, 0xac, 0x58, 0xcd}
	ERC1155InterfaceId = [4]byte{0xd9, 0xb6, 0x7a, 0x26}
	Addr1              = common.Address{0x1}
	Addr2              = common.Address{0x2}
	Addr3              = common.Address{0x3}
)

func init() {
	erc165 := "[{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
	ERC165ABI, _ = abi.JSON(strings.NewReader(erc165))
}

// SupportsInterface 查询给定合约是否支持interfaceId
func SupportsInterface(caller bind.ContractCaller, address common.Address, interfaceId [4]byte) (bool, error) {
	var out []interface{}

	err := Call(ERC165ABI, address, caller, DefaultOpts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err
}

func IsERC165(caller bind.ContractCaller, address common.Address) (bool, error) {
	support, err := SupportsInterface(caller, address, ERC165InterfaceId)
	err = FilterContractErr(err)
	if !support || err != nil {
		return false, err
	}
	support, err = SupportsInterface(caller, address, NotInterfaceId)
	err = FilterContractErr(err)
	return !support, err
}

func IsERC721(caller bind.ContractCaller, address common.Address) (bool, error) {
	support, err := SupportsInterface(caller, address, ERC721InterfaceId)
	err = FilterContractErr(err)
	return support, err
}

func IsERC1155(caller bind.ContractCaller, address common.Address) (bool, error) {
	support, err := SupportsInterface(caller, address, ERC1155InterfaceId)
	err = FilterContractErr(err)
	return support, err
}

func IsERC20(caller bind.ContractCaller, address common.Address) (ok bool, err error) {
	defer func() {
		err = FilterContractErr(err)
	}()
	totalSupply, err := TotalSupply(caller, address)
	if err != nil || totalSupply == nil {
		return false, err
	}
	balance, err := BalanceOf(caller, address, Addr1)
	if err != nil || balance == nil {
		return false, err
	}
	allowance, err := Allowance(caller, address, Addr1, Addr2)
	if err != nil || allowance == nil {
		return false, err
	}
	ok, err = Transfer(caller, &bind.CallOpts{From: Addr1}, address, Addr2, common.Big0)
	if err != nil || !ok {
		return false, err
	}
	ok, err = TransferFrom(caller, &bind.CallOpts{From: Addr1}, address, Addr2, Addr3, common.Big0)
	if err != nil || !ok {
		return false, err
	}
	ok, err = Approve(caller, &bind.CallOpts{From: Addr1}, address, Addr2, common.Big1)
	if err != nil || !ok {
		return false, err
	}
	return true, nil
}
