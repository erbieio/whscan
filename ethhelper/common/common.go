package common

import (
	"errors"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"strconv"
)

var (
	client *rpc.Client
)

func init() {
	client, _ = rpc.Dial(MainPoint)
	if client == nil {
		log.Println("rpc.Dial err")
		return
	}
}

func Client() *rpc.Client {
	return client
}

func BalanceOf(addr string) (Balance int64, err error) {
	var balance string
	err = client.Call(&balance, "eth_getBalance", addr, "latest")
	if err != nil {
		return -1, err
	}
	Balance, _ = strconv.ParseInt(balance, 0, 64)
	return Balance, nil
}


func GetBlock(number string) (Balance Block, err error) {
	var block Block
	err = client.Call(&block, "eth_getBlockByNumber", number, true)
	if err != nil {
		return block, err
	}
	return block, nil
}


func TransactionCount(addr string) (count int64, err error) {
	var c string
	err = client.Call(&c, "eth_getTransactionCount", addr, "latest")
	if err != nil {
		return -1, err
	}
	count, _ = strconv.ParseInt(c, 0, 64)
	return count, nil
}


func TransactionReceipt(txHash string) (ret Receipt, err error) {
	err = client.Call(&ret, "eth_getTransactionReceipt", txHash)
	if err != nil {
		return Receipt{}, err
	}
	return ret, nil
}


func ETHCall(params CallParamTemp) (ret string, err error) {
	err = client.Call(&ret, "eth_call", params, "latest")
	if err != nil {
		return "", err
	}
	return ret, nil
}

func GetLogs(filter LogFilter) (ret []Log, err error) {
	err = client.Call(&ret, "eth_getLogs", filter)
	if err != nil {
		return nil, err
	}
	return ret, nil
}


func ValidateSign(signHash, originData string) bool {
	return true
}

func SendRawTransaction(rawTransaction string) error {
	var ret string
	if err := client.Call(&ret, "eth_sendRawTransaction", rawTransaction); err != nil {
		return errors.New("Call failed:" + err.Error())
	}
	return nil
}
