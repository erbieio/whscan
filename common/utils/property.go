package utils

import (
	"context"
	"strings"

	"server/common/model"
	"server/common/types"
)

var (
	notInterfaceId     = "0xffffffff"
	erc165InterfaceId  = "0x01ffc9a7"
	erc721InterfaceId  = "0x80ac58cd"
	erc1155InterfaceId = "0xd9b67a26"

	Addr1 = "0x0000000000000000000000000000000000000001"
	Addr2 = "0x0000000000000000000000000000000000000002"
	Big0  = "0x0000000000000000000000000000000000000000000000000000000000000000"
)

func SetProperty(c ContractClient, ctx context.Context, number string, account *model.Account) (err error) {
	name, err := Name(c, ctx, number, account.Address)
	if err == nil {
		account.Name = &name
	}
	if err = filterContractErr(err); err != nil {
		return
	}
	symbol, err := Symbol(c, ctx, number, account.Address)
	if err == nil {
		account.Symbol = &symbol
	}
	if err = filterContractErr(err); err != nil {
		return
	}
	ok, err := IsERC165(c, ctx, number, account.Address)
	if err != nil {
		return
	}
	if !ok {
		ok, err = IsERC20(c, ctx, number, account.Address)
		if ok {
			account.Type = new(types.ContractType)
			*account.Type = types.ERC20
		}
		return
	}
	account.Type = new(types.ContractType)
	ok, err = IsERC721(c, ctx, number, account.Address)
	if err != nil {
		return
	}
	if ok {
		*account.Type = types.ERC721
		return
	}
	ok, err = IsERC1155(c, ctx, number, account.Address)
	if err != nil {
		return
	}
	if ok {
		*account.Type = types.ERC1155
		return
	}
	*account.Type = types.ERC165
	return
}

func IsERC165(c ContractClient, ctx context.Context, number, address any) (bool, error) {
	support, err := SupportsInterface(c, ctx, number, address, erc165InterfaceId)
	if !support || err != nil {
		return false, filterContractErr(err)
	}
	support, err = SupportsInterface(c, ctx, number, address, notInterfaceId)
	return !support, filterContractErr(err)
}

func IsERC721(c ContractClient, ctx context.Context, number, address any) (bool, error) {
	support, err := SupportsInterface(c, ctx, number, address, erc721InterfaceId)
	return support, filterContractErr(err)
}

func IsERC1155(c ContractClient, ctx context.Context, number, address any) (bool, error) {
	support, err := SupportsInterface(c, ctx, number, address, erc1155InterfaceId)
	return support, filterContractErr(err)
}

func IsERC20(c ContractClient, ctx context.Context, number, address any) (bool, error) {
	_, err := TotalSupply(c, ctx, number, address)
	if err != nil {
		return false, filterContractErr(err)
	}
	_, err = BalanceOf(c, ctx, number, address, Addr1)
	if err != nil {
		return false, filterContractErr(err)
	}
	_, err = Allowance(c, ctx, number, address, Addr1, Addr2)
	if err != nil {
		return false, filterContractErr(err)
	}
	_, err = Transfer(c, ctx, number, address, Addr2, Big0)
	if err != nil {
		return false, filterContractErr(err)
	}
	_, err = TransferFrom(c, ctx, number, address, Addr1, Addr2, Big0)
	if err != nil {
		return false, filterContractErr(err)
	}
	ok, err := Approve(c, ctx, number, address, Addr2, Big0)
	if !ok || err != nil {
		return false, filterContractErr(err)
	}
	return true, nil
}

func IsERC20_2(data string) bool {
	totalSupplySelector := "18160ddd"
	transferSelector := "a9059cbb"

	if strings.Contains(data, totalSupplySelector) &&
		strings.Contains(data, transferSelector) {
		return true
	}

	return false
}

func IsDelegateContract(log *model.EventLog) bool {
	delegateContractId := types.Hash("0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b")
	if log.Topics[0] == delegateContractId {
		return true
	}

	return false
}

func IsDelegateContract_2(topic0 string) bool {
	delegateContractId := "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b"
	if topic0 == delegateContractId {
		return true
	}

	return false
}

// filterContractErr Filter out errors except network connections
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

func GetContractType(c ContractClient, ctx context.Context, data string, number, address any) string {
	ok, err := IsERC165(c, ctx, number, address)
	if err != nil {
		return "ERCOther"
	}
	if !ok {
		ok = IsERC20_2(data)
		if ok {
			return "ERC20"
		}
		return "ERCOther"
	}

	ok, err = IsERC721(c, ctx, number, address)
	if err != nil {
		return "ERCOther"
	}
	if ok {
		return "ERC721"
	}
	ok, err = IsERC1155(c, ctx, number, address)
	if err != nil {
		return "ERCOther"
	}
	if ok {
		return "ERC1155"
	}
	return "ERC165"
}
