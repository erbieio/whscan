package node

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

var NotFound = fmt.Errorf("not found")

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

func (c *Client) CallContract(ctx context.Context, msg map[string]interface{}, blockNumber *types.BigInt) (types.Data, error) {
	var hex types.Data
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

// TraceTransaction implements the debug_traceTransaction interface to obtain transaction execution details (the reexec parameter can control the traced block depth)
func (c *Client) TraceTransaction(ctx context.Context, txHash types.Hash, options map[string]interface{}) (*ExecutionResult, error) {
	var r *ExecutionResult
	err := c.CallContext(ctx, &r, "debug_traceTransaction", txHash, options)
	if err == nil && r == nil {
		return nil, NotFound
	}
	return r, err
}

// TraceBlockByNumber implements the debug_traceBlockByNumber interface to obtain block execution details (the reexec parameter can control the traced block depth)
func (c *Client) TraceBlockByNumber(ctx context.Context, blockNumber types.Uint64, options map[string]interface{}) ([]*ExecutionResult, error) {
	var r []*ExecutionResult
	err := c.CallContext(ctx, &r, "debug_traceBlockByNumber", blockNumber.Hex(), options)
	if err == nil && r == nil {
		return nil, NotFound
	}
	return r, err
}

// GetInternalTx Get transaction internal call details
func (c *Client) GetInternalTx(ctx context.Context, tx *model.Transaction) (itx []*model.InternalTx, err error) {
	execRet, err := c.TraceTransaction(ctx, tx.Hash, map[string]interface{}{
		"disableStorage": true,
		"disableMemory":  true,
	})
	if err != nil {
		return
	}
	return c.decodeInternalTxs(ctx, execRet.StructLogs, tx)
}

// GetInternalTx Get transaction internal call details
func (c *Client) decodeInternalTxs(ctx context.Context, logs []StructLogRes, tx *model.Transaction) (itx []*model.InternalTx, err error) {
	to, number, txHash := new(types.Address), tx.BlockNumber, tx.Hash
	if tx.To != nil {
		*to = *tx.To
	} else {
		*to = *tx.ContractAddress
	}
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
			err = c.CallContext(ctx, &value, "eth_getBalance", tmp, number.Hex())
			if err != nil {
				return nil, err
			}
		case "create", "create2":
			checkDepth(&callers, log.Depth, to)
			value, to = utils.HexToBigInt(stack[len(stack)-1][2:]), new(types.Address)
			// The created address needs to be obtained in the first command stack after the return is created
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

func (c *Client) IsDebug() bool {
	return c.Call(&struct{}{}, "debug_gcStats") == nil
}

func (c *Client) IsWormholes() bool {
	return c.Call(&struct{}{}, "eth_getAccountInfo", "0x0000000000000000000000000000000000000000", "0x0") == nil
}

type Epoch struct {
	Dir        string   `json:"dir"`
	StartIndex uint64   `json:"start_index"`
	Royalty    uint32   `json:"royalty"`
	Creator    string   `json:"creator"`
	Address    string   `json:"address"` //Exchange address
	VoteWeight *big.Int `json:"vote_weight"`
}

func (c *Client) GetEpoch(number string) (rewards Epoch, err error) {
	err = c.Call(&rewards, "eth_getNominatedNFTInfo", number)
	return
}

type Reward struct {
	Address      string   `json:"Address"`
	Proxy        string   `json:"Proxy"`
	NFTAddress   *string  `json:"NftAddress"`
	RewardAmount *big.Int `json:"RewardAmount"`
}

func (c *Client) GetReward(number string) (rewards []*Reward, err error) {
	err = c.Call(&rewards, "eth_getBlockBeneficiaryAddressByNumber", number, true)
	if err != nil {
		return
	}
	validators := struct {
		Validators []*struct {
			Addr  string `json:"Addr"`
			Proxy string `json:"Proxy"`
		} `json:"Validators"`
	}{}
	err = c.Call(&validators, "eth_getValidator", number)
	if err != nil {
		return
	}
	for _, validator := range validators.Validators {
		for _, reward := range rewards {
			if validator.Addr == reward.Address {
				reward.Proxy = validator.Proxy
			}
		}
	}
	for _, reward := range rewards {
		if reward.Proxy == "" {
			reward.Proxy = reward.Address
		}
	}
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
	// todo because the chain is wrong, convert the MetaURL
	if len(snft.MetaURL) == 95 {
		snft.MetaURL = snft.MetaURL[0:53] + strings.ToLower(snft.MetaURL[91:93])
	}
	return
}

func (c *Client) GetPledge(addr, number string) (string, error) {
	pledge := struct {
		PledgedBalance *big.Int
	}{}
	err := c.Call(&pledge, "eth_getAccountInfo", addr, number)
	if pledge.PledgedBalance != nil {
		return pledge.PledgedBalance.Text(10), err
	} else {
		return "0", err
	}
}
