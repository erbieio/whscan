package extra

import (
	"context"
	"math/big"
	"strconv"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"server/common/types"
	"server/common/utils"
	"server/conf"
	"server/node"
)

var (
	client     *node.Client          //以太坊RPC客户端
	prv        *secp256k1.PrivateKey //私钥对象
	addr       types.Address         //私钥对应地址，持有大量测试ERB
	chainId    *big.Int              //链ID
	amount     *big.Int              //发送测试币的数量
	erbPayAddr string                //erbPay合约地址
)

func init() {
	var err error
	client, err = node.Dial(conf.ChainUrl)
	if err != nil {
		panic(err)
	}
	prv = conf.PrivateKey
	addr = utils.PubkeyToAddress(prv.PubKey())
	chainId = big.NewInt(conf.ChainId)
	amount = conf.Amount
	erbPayAddr = conf.ERBPay
}

// SendErb 发送测试ERB
func SendErb(to string, ctx context.Context) error {
	nonce, err := client.PendingNonceAt(ctx, addr)
	if err != nil {
		return err
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	tx := utils.NewTx(nonce, types.Address(to), amount, 21000, gasPrice, nil)
	rawTx, err := utils.SignTx(tx, chainId, prv)
	if err != nil {
		return err
	}

	return client.CallContext(ctx, nil, "eth_sendRawTransaction", rawTx)
}

func ExchangerAuth(addr string) (status uint64, flag bool, balance string, err error) {
	// 查询交易所状态
	account := struct {
		ExchangerFlag    bool
		ExchangerBalance *big.Int
	}{}
	err = client.Call(&account, "eth_getAccountInfo", addr, "latest")
	if err != nil {
		return
	}
	flag = account.ExchangerFlag
	balance = account.ExchangerBalance.String()

	// 调用ERBPay合约查询状态
	var hexRet string
	msg := map[string]interface{}{"to": erbPayAddr, "data": "0x4b165090" + "000000000000000000000000" + addr[2:]}
	err = client.Call(&hexRet, "eth_call", msg, "latest")
	if err != nil {
		return
	}
	if len(hexRet) > 2 {
		status, err = strconv.ParseUint(hexRet[2:], 16, 64)
	}
	return
}
