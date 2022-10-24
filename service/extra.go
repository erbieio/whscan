package service

import (
	"context"
	"math/big"
	"strconv"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"gorm.io/gorm/clause"
	"server/common/model"
	"server/common/types"
	"server/common/utils"
	"server/conf"
	"server/node"
)

var (
	client     *node.Client          //Ethereum RPC client
	prv        *secp256k1.PrivateKey //Private key object
	addr       types.Address         //The address corresponding to the private key holds a large number of test ERBs
	chainId    *big.Int              //Chain ID
	amount     *big.Int              //The amount of test coins sent
	erbPayAddr string                //erbPay contract address
)

func init() {
	var err error
	client, err = node.Dial(conf.ChainUrl)
	if err != nil {
		panic(err)
	}
	prv = conf.PrivateKey
	addr = utils.PubkeyToAddress(prv.PubKey())
	id, err := client.ChainId()
	chainId = new(big.Int).SetUint64(uint64(id))
	if err != nil {
		panic(err)
	}
	amount = conf.Amount
	erbPayAddr = conf.ERBPay
}

// SendErb sends the test ERB
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
	// Query exchange status
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

	// Call the ERBPay contract to query the status
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

func SaveSubscription(email string) error {
	return DB.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&model.Subscription{Email: email}).Error
}

func FetchSubscriptions(page, size int) (res []model.Subscription, err error) {
	err = DB.Offset((page - 1) * size).Limit(size).Find(&res).Error
	return
}
