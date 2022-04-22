package service

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"server/common/utils"
	"server/conf"
	"server/ethclient"
)

func SendErbForFaucet(to string) error {
	client, err := ethclient.Dial(conf.ChainUrl)
	if err != nil {
		return err
	}
	from := utils.PubkeyToAddress(conf.PrivateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	tx := types.NewTransaction(nonce, common.HexToAddress(to), conf.Amount, 21000, gasPrice, nil)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(conf.ChainId)), conf.PrivateKey)
	if err != nil {
		return err
	}

	return client.SendTransaction(context.Background(), signedTx)
}

type AuthRes struct {
	Status           uint64 `json:"status"`
	ExchangerFlag    bool   `json:"exchanger_flag"`
	ExchangerBalance string `json:"exchanger_balance"`
}

func ExchangerAuth(addr string) (res AuthRes, err error) {
	client, err := ethclient.Dial(conf.ChainUrl)
	if err != nil {
		return
	}
	exchanger, err := client.GetExchanger(addr)
	if err != nil {
		return
	}
	res.Status, err = checkAuth(client, addr)
	res.ExchangerFlag = exchanger.ExchangerFlag
	res.ExchangerBalance = exchanger.ExchangerBalance.String()
	return
}
