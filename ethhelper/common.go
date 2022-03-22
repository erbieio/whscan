package ethhelper

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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

func GetBlock(number string) (block Block, err error) {
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
func CheckAuth(addr string) (uint64, error) {
	var tmp big.Int
	payload := make([]byte, 36)

	tmp.SetString(common.CheckAuthHash, 0)
	copy(payload[:4], tmp.Bytes())
	tmp.SetString(addr, 0)
	copy(payload[36-len(tmp.Bytes()):], tmp.Bytes())
	params := CallParamTemp{To: common.ERBPayAddress, Data: "0x" + hex.EncodeToString(payload)}
	jsonData, err := json.Marshal(params)
	if err != nil {
		return 0, errors.New("Umarshal failed:" + err.Error() + string(jsonData))
	}
	fmt.Println(string(jsonData))
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

func SendErbForFaucet(toAddr string) error {
	client, err := ethclient.Dial(common.MainPoint)
	if err != nil {
		log.Println(err)
	}

	privateKey, err := crypto.HexToECDSA(common.SendErbPrivateKey)
	if err != nil {
		log.Println(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("cannot assert type: publicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	value, _ := new(big.Int).SetString("10000000000000000000000", 0) // in wei (100 erb)
	gasLimit := uint64(500000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	toAddress := common2.HexToAddress(toAddr)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
	return nil
}
