package service

import (
	"database/sql"
	"math/big"
	"strconv"

	"gorm.io/gorm"
	"server/conf"
	"server/node"
)

var (
	client     *node.Client //Ethereum RPC client
	erbPayAddr string       //erbPay contract address
)

func init() {
	var err error
	client, err = node.Dial(conf.ChainUrl)
	if err != nil {
		panic(err)
	}
	erbPayAddr = conf.ERBPay
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

func ExecSql(sqlStr string) (result []map[string]any, err error) {
	err = DB.Transaction(func(db *gorm.DB) error {
		return db.Raw(sqlStr).Scan(&result).Error
	}, &sql.TxOptions{ReadOnly: true})
	return
}
