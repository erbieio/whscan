package ethhelper

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
	"os/exec"
	"path/filepath"
	"server/ethhelper/common"
	"strconv"
)

var (
	client *rpc.Client
)

func init() {
	client, _ = rpc.Dial(common.MainPoint)
	if client == nil {
		log.Println("rpc.Dial err")
		return
	}
}

func GenCreateNftSign(contract, owner, metaUrl, tokenId, amount, royalty string) (error, string) {
	jsFile, err := filepath.Abs("ethhelper/jsassist/gen_nft_sign.js")
	if err != nil {
		return err, ""
	}
	cmd := exec.Command("node", jsFile, contract, owner, tokenId, amount, royalty, "")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return err, ""
	}
	fmt.Println("GenCreateNftSign args:", cmd.Args)
	if err = cmd.Wait(); err != nil {
		return err, ""
	}
	fmt.Println("GenCreateNftSign out:", out.String())
	return nil, out.String()
}

func BalanceOf(addr string) (balance string, err error) {
	err = client.Call(&balance, "eth_getBalance", addr, "latest")
	if err != nil {
		return "0", err
	}
	return balance, nil
}

func GetBlock(number string) (Balance Block, err error) {
	var block Block
	err = client.Call(&block, "eth_getBlockByNumber", number, true)
	if err != nil {
		return block, err
	}
	return block, nil
}
func GetBlockNumber() (num uint64, err error) {
	var ret string
	err = client.Call(&ret, "eth_blockNumber")
	if err != nil {
		return 0, err
	}
	b, _ := new(big.Int).SetString(ret, 0)
	return b.Uint64(), nil
}

func Decimals(contract string) (uint64, error) {
	var tmp big.Int
	payload := make([]byte, 4)

	tmp.SetString("0x313ce567", 0)
	copy(payload[:4], tmp.Bytes())
	params := CallParamTemp{To: contract, Data: "0x" + hex.EncodeToString(payload)}
	jsonData, err := json.Marshal(params)
	if err != nil {
		return 0, errors.New("Umarshal failed:" + err.Error() + string(jsonData))
	}

	var ret string
	if err = client.Call(&ret, "eth_call", params, "latest"); err != nil {
		return 0, errors.New("Call failed:" + err.Error())
	} else {
		tmp.SetString(ret, 0)
		return tmp.Uint64(), nil
	}
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
	err = client.Call(&ret, "eth_getLogs", filter, "latest")
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

func ContractCall(addr, input string) string {
	data := CallParamTemp{To: addr, Data: input}
	if ret, err := ETHCall(data); err == nil {
		return ret
	} else {
		return ""
	}
}

func CallViewFunc(contract, funcHash string) (string, error) {
	var tmp big.Int
	payload := make([]byte, 4)

	tmp.SetString(funcHash, 0)
	copy(payload[:4], tmp.Bytes())
	params := CallParamTemp{To: contract, Data: "0x" + hex.EncodeToString(payload)}
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", errors.New("Umarshal failed:" + err.Error() + string(jsonData))
	}

	var ret string
	if err = client.Call(&ret, "eth_call", params, "latest"); err != nil {
		return "", errors.New("Call failed:" + err.Error())
	} else {
		return ret, nil
	}
}
