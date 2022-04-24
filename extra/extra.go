package extra

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"server/conf"
)

var (
	client     *rpc.Client        //以太坊RPC客户端
	prv        *ecdsa.PrivateKey  //私钥对象
	addr       common.Address     //私钥对应地址，持有大量测试ERB
	signer     types.EIP155Signer //交易签名对象
	amount     *big.Int           //发送测试币的数量
	erbPayAddr string             //erbPay合约地址
)

func init() {
	var err error
	client, err = rpc.Dial(conf.ChainUrl)
	if err != nil {
		panic(err)
	}
	prv, err = crypto.HexToECDSA(conf.HexKey)
	if err != nil {
		panic(err)
	}
	addr = crypto.PubkeyToAddress(prv.PublicKey)
	signer = types.NewEIP155Signer(big.NewInt(conf.ChainId))
	amount = new(big.Int)
	amount.SetString(conf.AmountStr, 0)
	erbPayAddr = conf.ERBPay
}

// SendErb 发送测试ERB
func SendErb(to string, ctx context.Context) error {
	client := ethclient.NewClient(client)
	nonce, err := client.PendingNonceAt(ctx, addr)
	if err != nil {
		return err
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	tx := types.NewTransaction(nonce, common.HexToAddress(to), amount, 21000, gasPrice, nil)
	signedTx, err := types.SignTx(tx, signer, prv)
	if err != nil {
		return err
	}
	return client.SendTransaction(ctx, signedTx)
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
