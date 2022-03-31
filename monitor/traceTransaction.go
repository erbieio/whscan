package monitor

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	common2 "server/ethhelper/common"
	"server/log"
	"strings"
)

type EthClient struct {
	*ethclient.Client
	c *rpc.Client
}

func Dial(rawurl string) (*EthClient, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*EthClient, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *EthClient {
	return &EthClient{
		Client: ethclient.NewClient(c),
		c:      c,
	}
}

type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

type StructLogRes struct {
	Pc      uint64    `json:"pc"`
	Op      string    `json:"op"`
	Gas     uint64    `json:"gas"`
	GasCost uint64    `json:"gasCost"`
	Depth   int       `json:"depth"`
	Error   string    `json:"error,omitempty"`
	Stack   *[]string `json:"stack,omitempty"`
}

// TraceTransaction 继承ethClient并实现debug_traceTransaction接口，获取合约字节码执行详情（最近128区块内交易有效）
func (ec *EthClient) TraceTransaction(ctx context.Context, txHash common.Hash) (*ExecutionResult, error) {
	var r *ExecutionResult
	err := ec.c.CallContext(ctx, &r, "debug_traceTransaction", txHash, map[string]interface{}{"disableMemory": true, "disableStorage": true})
	if err == nil && r == nil {
		return nil, ethereum.NotFound
	}
	return r, err
}

type CallLog struct {
	Op       string `json:"op"`
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	GasLimit string `json:"gasLimit"`
}

func TraceTxInternalCall(txHash, from, to string) []CallLog {
	client, err := Dial(common2.MainPoint)
	if err != nil {
		log.Infof(err.Error())
	}

	ret, err := client.TraceTransactionCall(context.Background(), common.HexToHash(txHash), from, to)
	if err != nil {
		log.Infof(err.Error())
	}
	return ret
}

// TraceTransactionCall 获取交易内部调用详情
func (ec *EthClient) TraceTransactionCall(ctx context.Context, txHash common.Hash, from, toAddr string) ([]CallLog, error) {
	executionResult, err := ec.TraceTransaction(ctx, txHash)
	if err != nil {
		return nil, err
	}
	tx, _, err := ec.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, err
	}
	ret := []CallLog{{Op: "CALL", From: from, To: toAddr, Value: toDecimal(tx.Value().String(), 18), GasLimit: fmt.Sprintf("%v", tx.Gas())}}
	for _, log := range executionResult.StructLogs {
		stack := *log.Stack
		switch log.Op {
		case "CALL", "CALLCODE":
			to, value := stack[len(stack)-2], stack[len(stack)-3]
			if !strings.Contains(value, ".") {
				value = toDecimal(value, 18)
			}

			ret = append(ret, CallLog{Op: log.Op, To: common.HexToAddress(to).String(), Value: value, GasLimit: fmt.Sprintf("%v", log.Gas)})
		case "DELEGATECALL", "STATICCALL":
			to := stack[len(stack)-2]
			var value string
			if !strings.Contains(ret[len(ret)-1].Value, ".") {
				value = toDecimal(ret[len(ret)-1].Value, 18)
			} else {
				value = ret[len(ret)-1].Value
			}
			ret = append(ret, CallLog{Op: log.Op, To: common.HexToAddress(to).String(), GasLimit: fmt.Sprintf("%v", log.Gas), Value: value})
		default:
			continue
		}
		// 上一级的to就是这一级调用的from
		ret[len(ret)-1].From = ret[len(ret)-2].To
	}
	// 去掉交易本身的CALL调用
	return ret, nil
}
