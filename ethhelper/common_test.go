package ethhelper

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
)

func TestTxCount(t *testing.T) {
  b,e:=BalanceOf("0x10CEc672c6BB2f6782BEED65987E020902B7bD15")
  fmt.Println(b,e)
}

func TestTransactionReceipt(t *testing.T) {
	c, err := TransactionReceipt("0x6d1443e8b5682f94aca101a9455508ee4777c4c68b81752b9c1a04734cb0c919")
	fmt.Println(c, err)
}
func TestAdminList(t *testing.T) {
	c, err := IsOwnerOfNFT1155("0x86c02ffd61b0aca14ced6c3fefc4c832b58b246c", "0xa1e67a33e090afe696d7317e05c506d7687bb2e5", "2224819644905")
	fmt.Println(c, err)
}
func TestSendDealAuctionTx(t *testing.T) {
	SendDealAuctionTx("0x10CEc672c6BB2f6782BEED65987E020902B7bD15", "0x572bcAcB7ae32Db658C8dEe49e156d455Ad59eC8", "0x58C68d71F7E8063c25097d938e7857582D5a1c70", "34", "20000000000000000", "0x844c0fe2c4183f3b73931b9f165aaff1013fbaec7fc64c4c491e52e16e1d0f12475bbb36e0028a5948e6fd59d666e6e6b918d5cb2cf309f07b9d67fd4c8586f51c")
}
func TestGetBlock(t *testing.T) {
	var b big.Int
	b.SetString("100", 0)
	s := hex.EncodeToString(b.Bytes())
	fmt.Println(GetBlock("0x" + s))
}

var httpUrl = "http://192.168.1.235:8561"
/*
	以太坊交易发送
*/
func TestSendTx(t *testing.T) {
	SendErbForFaucet("0xCb7Bf52B5DCccA63C2Bf9b9775D1DbBE0295d37a")
}
