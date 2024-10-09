package utils

import (
	"context"
	"math/big"
	"server/node"
	"testing"
)

func createClient() *node.Client {
	var err error
	client := &node.Client{}
	if client, err = node.Dial("http://192.168.1.235:8560"); err != nil {
		return nil
	}
	//if client, err = node.Dial("http://192.168.84.240:8560"); err != nil {
	//	return nil
	//}
	//if client, err = node.Dial("https://mainnet.infura.io/v3/b6bf7d3508c941499b10025c0776eaf8"); err != nil {
	//	return nil
	//}

	return client
}

func TestName(t *testing.T) {
	client := createClient()
	name, err := Name(client, context.Background(), "0x100", "0x1f22e90d61d08c4b327139728902682ca2bcb042")
	//name, err := Name(client, context.Background(), "0x13D1F55", "0xdAC17F958D2ee523a2206206994597C13D831ec7 ")

	//data, err := hex.DecodeString(string("584c"))

	t.Log("name: ", name, "err: ", err)
}

func TestSymbol(t *testing.T) {
	client := createClient()
	symbol, err := Symbol(client, context.Background(), "0x100", "0x1f22e90d61d08c4b327139728902682ca2bcb042")
	//name, err := Name(client, context.Background(), "0x13D1F55", "0xdAC17F958D2ee523a2206206994597C13D831ec7 ")

	//data, err := hex.DecodeString(string("584c"))

	t.Log("symbol: ", symbol, "err: ", err)
}

func TestTotalSupply(t *testing.T) {
	client := createClient()
	//totalSupply, err := TotalSupply(client, context.Background(), "0x100", "0x1f22e90d61d08c4b327139728902682ca2bcb042")
	totalSupply, err := TotalSupply(client, context.Background(), "0x224c9f", "0x9c5e37716861a7e03976fb996228c00d31dd40ea")
	//name, err := Name(client, context.Background(), "0x13D1F55", "0xdAC17F958D2ee523a2206206994597C13D831ec7 ")

	//data, err := hex.DecodeString(string("584c"))
	total, _ := new(big.Int).SetString(totalSupply[2:], 16)
	total.Div(total, big.NewInt(1000000000000000000))

	t.Log("total: ", total, "err: ", err)
}

func TestIsERC20(t *testing.T) {
	client := createClient()
	flag, err := IsERC20(client, context.Background(), "0x100", "0x746c57d369849e73f27c7981331edfb8bcab7d89")
	//name, err := Name(client, context.Background(), "0x13D1F55", "0xdAC17F958D2ee523a2206206994597C13D831ec7 ")

	//data, err := hex.DecodeString(string("584c"))

	t.Log("flag: ", flag, "err: ", err)
}

func TestD(t *testing.T) {
	am, _ := new(big.Int).SetString("10bf9519174d5f9000", 16)
	t.Log(am)
}

func TestGetContractType(t *testing.T) {
	client := createClient()
	ty := GetContractType(client, context.Background(), "", "0x59a27", "0xce1cc1aa4e1cee97b13c9e650a7be66345d7d04f")
	t.Log(ty)
}

func TestGetTokenURI(t *testing.T) {
	client := createClient()
	ty, err := GetTokenURI(client, context.Background(), "0x59a65", "0xce1cc1aa4e1cee97b13c9e650a7be66345d7d04f", 1)
	t.Log(ty, err)
}

func TestGetOwnerOf(t *testing.T) {
	client := createClient()
	ty, err := GetOwnerOf(client, context.Background(), "0x59a65", "0xce1cc1aa4e1cee97b13c9e650a7be66345d7d04f", 1)
	t.Log(ty, err)
}

func TestBalanceOf1155(t *testing.T) {
	client := createClient()
	toQuantity, err := BalanceOf1155(client, context.Background(), "0x7668a", "0x03cff07122b8e82418bd9152763516f7141a2c39", "0x15b1049c7f8fb1b5f1cba8236eac0e77fbee4e66", 0)
	t.Log(toQuantity, err)
}
func TestGetTokenURI1155(t *testing.T) {
	client := createClient()
	ty, err := GetTokenURI1155(client, context.Background(), "0x72627", "0x03cff07122b8e82418bd9152763516f7141a2c39", 0)
	t.Log(ty, err)
}
