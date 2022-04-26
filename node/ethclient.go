package node

import (
	"context"
	"math/big"
	"strconv"
	"strings"

	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client struct {
	*RPC
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	rpc, err := NewRPC(rawurl)
	return &Client{rpc}, err
}

type Big big.Int

func (b *Big) UnmarshalJSON(input []byte) error {
	return (*big.Int)(b).UnmarshalJSON(input[1 : len(input)-1])
}

func (c *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	var hex Big
	if err := c.CallContext(ctx, &hex, "eth_gasPrice"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

func (c *Client) PendingNonceAt(ctx context.Context, account types.Address) (uint64, error) {
	var result string
	err := c.CallContext(ctx, &result, "eth_getTransactionCount", account, "pending")
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(result[2:], 16, 64)
}

func (c *Client) CallContract(ctx context.Context, msg map[string]interface{}, blockNumber *types.BigInt) (types.Bytes, error) {
	var hex types.Bytes
	err := c.CallContext(ctx, &hex, "eth_call", msg, toBlockNumArg(blockNumber))
	if err != nil {
		return "", err
	}
	return hex, nil
}

func toBlockNumArg(number *types.BigInt) string {
	if number == nil {
		return "latest"
	}
	if *number == "-1" {
		return "pending"
	}
	return number.Hex()
}

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
func (c *Client) TraceTransaction(ctx context.Context, txHash string, options map[string]interface{}) (*ExecutionResult, error) {
	var r *ExecutionResult
	err := c.CallContext(ctx, &r, "debug_traceTransaction", txHash, options)
	if err == nil && r == nil {
		return nil, NotFound
	}
	return r, err
}

// TraceBlockByNumber 实现debug_traceBlockByNumber接口，获取区块执行详情（reexec参数可以控制追溯的块深度）
func (c *Client) TraceBlockByNumber(ctx context.Context, blockNumber types.Uint64, options map[string]interface{}) ([]*ExecutionResult, error) {
	var r []*ExecutionResult
	err := c.CallContext(ctx, &r, "debug_traceBlockByNumber", blockNumber.Hex(), options)
	if err == nil && r == nil {
		return nil, NotFound
	}
	return r, err
}

// GetInternalTx 获取交易内部调用详情
func (c *Client) GetInternalTx(ctx context.Context, number types.Uint64, txHash types.Hash, to types.Address) (itx []*model.InternalTx, err error) {
	execRet, err := c.TraceTransaction(ctx, string(txHash), map[string]interface{}{
		"disableStorage": true,
		"disableMemory":  true,
	})
	if err != nil {
		return
	}

	return c.decodeInternalTxs(ctx, execRet.StructLogs, txHash, &to)
}

// GetInternalTx 获取交易内部调用详情
func (c *Client) decodeInternalTxs(ctx context.Context, logs []StructLogRes, txHash types.Hash, to *types.Address) (itx []*model.InternalTx, err error) {
	callers, createLogs := []*types.Address{to}, make([]*model.InternalTx, 0)
	checkDepth := func(callers *[]*types.Address, depth int, to *types.Address) {
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
				*createLog.To = types.Address((*nextLog.Stack)[len(*nextLog.Stack)-1])
				createLogs = createLogs[:len(createLogs)-1]
			}
		}
	}

	for i, log := range logs {
		stack, op, value := *log.Stack, strings.ToLower(log.Op), types.BigInt("0")
		switch op {
		case "call", "callcode":
			checkDepth(&callers, log.Depth, to)
			tmp := utils.HexToAddress(stack[len(stack)-2][2:])
			to, value = &tmp, utils.HexToBigInt(stack[len(stack)-3][2:])
		case "delegatecall":
			callers = append(callers, callers[len(callers)-1])
			tmp := utils.HexToAddress(stack[len(stack)-2][2:])
			to = &tmp
		case "staticcall":
			checkDepth(&callers, log.Depth, to)
			tmp := utils.HexToAddress(stack[len(stack)-2][2:])
			to = &tmp
		case "selfdestruct":
			checkDepth(&callers, log.Depth, to)
			tmp := utils.HexToAddress(stack[len(stack)-1][2:])
			to = &tmp
			setCreateAddr(i)
		case "create", "create2":
			checkDepth(&callers, log.Depth, to)
			value, to = utils.HexToBigInt(stack[len(stack)-1][2:]), new(types.Address)
			// 创建的地址需要等到创建return之后的第一个命令栈里面获取
			createLogs = append(createLogs, &model.InternalTx{
				Depth: types.Uint64(log.Depth),
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
			Depth:        types.Uint64(log.Depth),
			Op:           op,
			From:         callers[log.Depth-1],
			To:           to,
			Value:        value,
			GasLimit:     types.Uint64(log.Gas),
		})
	}
	return
}

type Reward struct {
	Address    string
	NfTAddress string
}

func (c *Client) GetReward(number string) (rewards []Reward, err error) {
	err = c.Call(&rewards, "eth_getBlockBeneficiaryAddressByNumber", number, true)
	return
}

type SNFT struct {
	Owner   string
	Creator string
	Royalty uint32
	MetaURL string
}

func (c *Client) GetSNFT(addr, number string) (snft SNFT, err error) {
	err = c.Call(&snft, "eth_getAccountInfo", addr, number)
	// todo 因为链错误，固定SNFT的MetaURL
	if len(snft.MetaURL) > 0 {
		snft.MetaURL = "/ipfs/QmeCPcX3rYguWqJYDmJ6D4qTQqd5asr8gYpwRcgw44WsS7/00"
	}
	return
}
