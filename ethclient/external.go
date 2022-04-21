package ethclient

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

type StructLogRes struct {
	Pc      uint64             `json:"pc"`
	Op      string             `json:"op"`
	Gas     uint64             `json:"gas"`
	GasCost uint64             `json:"gasCost"`
	Depth   int                `json:"depth"`
	Error   string             `json:"error,omitempty"`
	Stack   *[]string          `json:"stack,omitempty"`
	Memory  *[]string          `json:"memory,omitempty"`
	Storage *map[string]string `json:"storage,omitempty"`
}

// TraceTransaction 实现debug_traceTransaction接口，获取交易执行详情（reexec参数可以控制追溯的块深度）
func (ec *Client) TraceTransaction(ctx context.Context, txHash string, options map[string]interface{}) (*ExecutionResult, error) {
	var r *ExecutionResult
	err := ec.c.CallContext(ctx, &r, "debug_traceTransaction", txHash, options)
	if err == nil && r == nil {
		return nil, ethereum.NotFound
	}
	return r, err
}

// TraceBlockByNumber 实现debug_traceBlockByNumber接口，获取区块执行详情（reexec参数可以控制追溯的块深度）
func (ec *Client) TraceBlockByNumber(ctx context.Context, blockNumber uint64, options map[string]interface{}) ([]*ExecutionResult, error) {
	var r []*ExecutionResult
	err := ec.c.CallContext(ctx, &r, "debug_traceBlockByNumber", hexutil.EncodeUint64(blockNumber), options)
	if err == nil && r == nil {
		return nil, ethereum.NotFound
	}
	return r, err
}

func (ec *Client) Call(result interface{}, method string, args ...interface{}) error {
	ctx := context.Background()
	return ec.c.CallContext(ctx, result, method, args...)
}

type Reward struct {
	Address    string
	NfTAddress string
}

func (ec *Client) GetReward(number string) (rewards []Reward, err error) {
	err = ec.c.Call(&rewards, "eth_getBlockBeneficiaryAddressByNumber", number, true)
	return
}

type Exchanger struct {
	ExchangerFlag    bool
	ExchangerBalance big.Int
}

func (ec *Client) GetExchanger(addr string) (e Exchanger, err error) {
	err = ec.c.Call(&e, "eth_getAccountInfo", addr, "latest")
	return
}

type SNFT struct {
	Owner   string
	Creator string
	Royalty uint32
	MetaURL string
}

func (ec *Client) GetSNFT(addr, number string) (snft SNFT, err error) {
	err = ec.c.Call(&snft, "eth_getAccountInfo", addr, number)
	// todo 因为链错误，固定SNFT的MetaURL
	if len(snft.MetaURL) > 0 {
		snft.MetaURL = "/ipfs/QmeCPcX3rYguWqJYDmJ6D4qTQqd5asr8gYpwRcgw44WsS7/00"
	}
	return
}

// realMeatUrl 解析真正的metaUrl
func realMeatUrl(meta string) string {
	data, err := hexutil.Decode("0x" + meta)
	if err != nil {
		return ""
	}
	r := struct {
		Meta string `json:"meta"`
	}{}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return ""
	}
	return r.Meta
}

// hashMsg Wormholes链代码复制 go-ethereum/core/evm.go 330行
func hashMsg(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), string(data))
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	return hasher.Sum(nil), msg
}

// recoverAddress Wormholes链代码复制 go-ethereum/core/evm.go 338行
func recoverAddress(msg string, sigStr string) (string, error) {
	sigData := hexutil.MustDecode(sigStr)
	if len(sigData) != 65 {
		return common.Address{}.Hex(), fmt.Errorf("signature must be 65 bytes long")
	}
	if sigData[64] != 27 && sigData[64] != 28 {
		return common.Address{}.Hex(), fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sigData[64] -= 27
	hash, _ := hashMsg([]byte(msg))
	rpk, err := crypto.SigToPub(hash, sigData)
	if err != nil {
		return common.Address{}.Hex(), err
	}
	return strings.ToLower(crypto.PubkeyToAddress(*rpk).Hex()), nil
}
