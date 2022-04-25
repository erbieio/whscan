package ethclient

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strings"

	"server/common/model"
	. "server/common/types"
	"server/common/utils"
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
	err := ec.CallContext(ctx, &r, "debug_traceTransaction", txHash, options)
	if err == nil && r == nil {
		return nil, NotFound
	}
	return r, err
}

// TraceBlockByNumber 实现debug_traceBlockByNumber接口，获取区块执行详情（reexec参数可以控制追溯的块深度）
func (ec *Client) TraceBlockByNumber(ctx context.Context, blockNumber Uint64, options map[string]interface{}) ([]*ExecutionResult, error) {
	var r []*ExecutionResult
	err := ec.CallContext(ctx, &r, "debug_traceBlockByNumber", blockNumber.Hex(), options)
	if err == nil && r == nil {
		return nil, NotFound
	}
	return r, err
}

// GetInternalTx 获取交易内部调用详情
func (ec *Client) GetInternalTx(ctx context.Context, number Uint64, txHash Hash, to Address) (itx []*model.InternalTx, err error) {
	execRet, err := ec.TraceTransaction(ctx, string(txHash), map[string]interface{}{
		"disableStorage": true,
		"disableMemory":  true,
	})
	if err != nil {
		return
	}

	return ec.decodeInternalTxs(ctx, execRet.StructLogs, txHash, &to)
}

// GetInternalTx 获取交易内部调用详情
func (ec *Client) decodeInternalTxs(ctx context.Context, logs []StructLogRes, txHash Hash, to *Address) (itx []*model.InternalTx, err error) {
	callers, createLogs := []*Address{to}, make([]*model.InternalTx, 0)
	checkDepth := func(callers *[]*Address, depth int, to *Address) {
		if depth > len(*callers) {
			*callers = append(*callers, to)
		} else if depth < len(*callers) {
			*callers = (*callers)[:len(*callers)-1]
		}
	}
	setCreateAddr := func(i int) {
		if len(createLogs) > 0 {
			nextLog := logs[i+1]
			createLog := createLogs[len(createLogs)-1]
			if int(createLog.Depth) == nextLog.Depth {
				*createLog.To = Address((*nextLog.Stack)[len(*nextLog.Stack)-1])
				createLogs = createLogs[:len(createLogs)-1]
			}
		}
	}

	for i, log := range logs {
		stack, op, value := *log.Stack, strings.ToLower(log.Op), BigInt("0")
		switch op {
		case "call", "callcode":
			checkDepth(&callers, log.Depth, to)
			tmp := utils.HexToAddress(stack[len(stack)-2])
			to, value = &tmp, utils.HexToBigInt(stack[len(stack)-3][2:])
		case "delegatecall":
			callers = append(callers, callers[len(callers)-1])
			tmp := utils.HexToAddress(stack[len(stack)-2])
			to = &tmp
		case "staticcall":
			checkDepth(&callers, log.Depth, to)
			tmp := utils.HexToAddress(stack[len(stack)-2])
			to = &tmp
		case "selfdestruct":
			checkDepth(&callers, log.Depth, to)
			tmp := utils.HexToAddress(stack[len(stack)-1])
			to = &tmp
			setCreateAddr(i)
		case "create", "create2":
			checkDepth(&callers, log.Depth, to)
			value, to = utils.HexToBigInt(stack[len(stack)-1][2:]), new(Address)
			// 创建的地址需要等到创建return之后的第一个命令栈里面获取
			createLogs = append(createLogs, &model.InternalTx{
				Depth: Uint64(log.Depth),
				To:    to,
			})
		case "return", "revert":
			setCreateAddr(i)
			continue
		default:
			continue
		}
		itx = append(itx, &model.InternalTx{
			ParentTxHash: txHash,
			Depth:        Uint64(log.Depth),
			Op:           op,
			From:         callers[log.Depth-1],
			To:           to,
			Value:        value,
			GasLimit:     Uint64(log.Gas),
		})
	}
	return
}

type Reward struct {
	Address    string
	NfTAddress string
}

func (ec *Client) GetReward(number string) (rewards []Reward, err error) {
	err = ec.Call(&rewards, "eth_getBlockBeneficiaryAddressByNumber", number, true)
	return
}

type SNFT struct {
	Owner   string
	Creator string
	Royalty uint32
	MetaURL string
}

func (ec *Client) GetSNFT(addr, number string) (snft SNFT, err error) {
	err = ec.Call(&snft, "eth_getAccountInfo", addr, number)
	// todo 因为链错误，固定SNFT的MetaURL
	if len(snft.MetaURL) > 0 {
		snft.MetaURL = "/ipfs/QmeCPcX3rYguWqJYDmJ6D4qTQqd5asr8gYpwRcgw44WsS7/00"
	}
	return
}

// realMeatUrl 解析真正的metaUrl
func realMeatUrl(meta string) string {
	data, err := hex.DecodeString(meta)
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
