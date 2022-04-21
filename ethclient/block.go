package ethclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	. "server/common/types"
	"server/common/utils"
	"server/model"
)

type Block struct {
	*model.Block
	CacheTxs         []model.Transaction
	CacheInternalTxs []*model.InternalTx
	CacheUncles      []*model.Uncle
	CacheLogs        []*model.Log
	CacheAccounts    map[Bytes20]*model.Account
	CacheContracts   map[Bytes20]*model.Contract
}

func (ec *Client) GetBlock(ctx context.Context, number uint64) (*Block, error) {
	var raw json.RawMessage
	// 获取区块及其交易
	err := ec.c.CallContext(ctx, &raw, "eth_getBlockByNumber", hexutil.EncodeUint64(number), false)
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}

	// Quick-verify transaction and uncle lists. This mostly helps with debugging the server.
	if string(block.Sha3Uncles) == types.EmptyUncleHash.String() && len(block.UncleHashes) > 0 {
		return nil, fmt.Errorf("server returned non-empty uncle list but block header indicates no uncles")
	}
	if string(block.Sha3Uncles) != types.EmptyUncleHash.String() && len(block.UncleHashes) == 0 {
		return nil, fmt.Errorf("server returned empty uncle list but block header indicates uncles")
	}
	if string(block.TransactionsRoot) == types.EmptyRootHash.String() && len(block.Transactions) > 0 {
		return nil, fmt.Errorf("server returned non-empty transaction list but block header indicates no transactions")
	}
	if string(block.TransactionsRoot) != types.EmptyRootHash.String() && len(block.Transactions) == 0 {
		return nil, fmt.Errorf("server returned empty transaction list but block header indicates transactions")
	}

	if totalTransaction := len(block.Transactions); totalTransaction > 0 {
		block.TotalTransaction = Uint64(totalTransaction)
		block.CacheTxs = make([]model.Transaction, totalTransaction)
		// 获取交易和收据
		reqs := make([]rpc.BatchElem, totalTransaction*2)
		for i, hash := range block.Transactions {
			reqs[i] = rpc.BatchElem{
				Method: "eth_getTransactionByHash",
				Args:   []interface{}{hash},
				Result: &block.CacheTxs[i],
			}
			reqs[i+totalTransaction] = rpc.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []interface{}{hash},
				Result: &block.CacheTxs[i].Receipt,
			}
		}
		if err := ec.c.BatchCallContext(ctx, reqs); err != nil {
			return nil, err
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, reqs[i].Error
			}
		}
		// 获取收据logs,只能根据区块哈希查，区块高度会查到
		err := ec.c.CallContext(ctx, &block.CacheLogs, "eth_getLogs", map[string]interface{}{"blockHash": block.Hash})
		if err != nil {
			return nil, err
		}
		// 获取解析内部交易
		//for _, tx := range block.CacheTxs {
		//	to := tx.To
		//	if to == nil {
		//		to = tx.ContractAddress
		//	}
		//	internalTxs, err := ec.GetInternalTx(ctx, number, tx.Hash, *to)
		//	if err != nil {
		//		return nil, err
		//	}
		//	block.CacheInternalTxs = append(block.CacheInternalTxs, internalTxs...)
		//}
	}

	// 获取叔块
	block.UnclesCount = Uint64(len(block.UncleHashes))
	if block.UnclesCount > 0 {
		block.CacheUncles = make([]*model.Uncle, block.UnclesCount)
		reqs := make([]rpc.BatchElem, block.UnclesCount)
		for i := range reqs {
			reqs[i] = rpc.BatchElem{
				Method: "eth_getUncleByBlockHashAndIndex",
				Args:   []interface{}{block.Hash, hexutil.Uint64(i)},
				Result: &block.CacheUncles[i],
			}
		}
		if err := ec.c.BatchCallContext(ctx, reqs); err != nil {
			return nil, err
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, reqs[i].Error
			}
		}
	}

	// 解析相关帐户和创建的合约
	block.CacheAccounts = make(map[Bytes20]*model.Account)
	block.CacheContracts = make(map[Bytes20]*model.Contract)
	block.CacheAccounts[block.Miner] = &model.Account{Address: block.Miner}
	if len(block.CacheTxs) > 0 {
		for _, tx := range block.CacheTxs {
			block.CacheAccounts[tx.From] = &model.Account{Address: tx.From}
			if tx.To != nil {
				block.CacheAccounts[*tx.To] = &model.Account{Address: *tx.To}
			}
			if tx.ContractAddress != nil {
				block.CacheAccounts[*tx.ContractAddress] = &model.Account{Address: *tx.ContractAddress}
				block.CacheContracts[*tx.ContractAddress] = &model.Contract{Address: *tx.ContractAddress, Creator: tx.From, CreatedTx: tx.Hash}
			}
		}
	}

	// 获取帐户属性
	reqs := make([]rpc.BatchElem, 0, 2*len(block.CacheAccounts))
	for _, account := range block.CacheAccounts {
		reqs = append(reqs, rpc.BatchElem{
			Method: "eth_getBalance",
			Args:   []interface{}{account.Address, toBlockNumArg(nil)},
			Result: &account.Balance,
		})
		reqs = append(reqs, rpc.BatchElem{
			Method: "eth_getTransactionCount",
			Args:   []interface{}{account.Address, toBlockNumArg(nil)},
			Result: &account.Nonce,
		})
	}
	if err := ec.c.BatchCallContext(ctx, reqs); err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].Error != nil {
			return nil, reqs[i].Error
		}
	}

	// 获取合约属性
	if len(block.CacheContracts) > 0 {
		reqs := make([]rpc.BatchElem, 0, len(block.CacheContracts))
		for _, contract := range block.CacheContracts {
			reqs = append(reqs, rpc.BatchElem{
				Method: "eth_getCode",
				Args:   []interface{}{contract.Address, toBlockNumArg(nil)},
				Result: &contract.Code,
			})
		}
		if err := ec.c.BatchCallContext(ctx, reqs); err != nil {
			return nil, err
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, reqs[i].Error
			}
		}

		for _, contract := range block.CacheContracts {
			hash := crypto.Keccak256Hash(hexutil.MustDecode(contract.Code)).String()
			block.CacheAccounts[contract.Address].CodeHash = (*Bytes32)(&hash)
			if len(contract.Code) > 2 {
				contract.ERC, err = ec.GetERC(common.HexToAddress(string(contract.Address)))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &block, nil
}

// GetERC 获取合约类型，合约调用的错误将不会返回且合约将视为无类型的合约
func (ec *Client) GetERC(address common.Address) (ERC, error) {
	ok, err := utils.IsERC165(ec, address)
	if err != nil {
		return NONE, err
	}
	if !ok {
		ok, err = utils.IsERC20(ec, address)
		if ok && err == nil {
			return ERC20, nil
		} else {
			return NONE, err
		}
	}
	ok, err = utils.IsERC721(ec, address)
	if err != nil {
		return NONE, err
	}
	if ok {
		return ERC721, nil
	}
	ok, err = utils.IsERC1155(ec, address)
	if err != nil {
		return NONE, err
	}
	if ok {
		return ERC1155, nil
	}
	return ERC165, nil
}

// GetInternalTx 获取交易内部调用详情
func (ec *Client) GetInternalTx(ctx context.Context, number uint64, txHash Bytes32, to Bytes20) (itx []*model.InternalTx, err error) {
	execRet, err := ec.TraceTransaction(ctx, string(txHash), map[string]interface{}{
		"disableStorage": true,
		"disableMemory":  true,
	})
	if err != nil {
		return
	}

	return ec.decodeInternalTxs(ctx, execRet.StructLogs, number, txHash, &to)
}

// GetInternalTx 获取交易内部调用详情
func (ec *Client) decodeInternalTxs(ctx context.Context, logs []StructLogRes, number uint64, txHash Bytes32, to *Bytes20) (itx []*model.InternalTx, err error) {
	callers, createLogs := []*Bytes20{to}, make([]*model.InternalTx, 0)
	checkDepth := func(callers *[]*Bytes20, depth int, to *Bytes20) {
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
				*createLog.To = Bytes20((*nextLog.Stack)[len(*nextLog.Stack)-1])
				createLogs = createLogs[:len(createLogs)-1]
			}
		}
	}

	for i, log := range logs {
		stack, op, value := *log.Stack, strings.ToLower(log.Op), "0x0"
		switch op {
		case "call", "callcode":
			checkDepth(&callers, log.Depth, to)
			to, value = (*Bytes20)(&stack[len(stack)-2]), stack[len(stack)-3]
		case "delegatecall":
			callers = append(callers, callers[len(callers)-1])
			to, value = (*Bytes20)(&stack[len(stack)-2]), "0x0"
		case "staticcall":
			checkDepth(&callers, log.Depth, to)
			to, value = (*Bytes20)(&stack[len(stack)-2]), "0x0"
		case "selfdestruct":
			checkDepth(&callers, log.Depth, to)
			to = (*Bytes20)(&stack[len(stack)-1])
			setCreateAddr(i)
			err = ec.c.CallContext(ctx, &value, "eth_getBalance", common.HexToAddress(string(*callers[log.Depth-1])), hexutil.EncodeUint64(number))
			if err != nil {
				return nil, err
			}
		case "create", "create2":
			checkDepth(&callers, log.Depth, to)
			value, to = stack[len(stack)-1], new(Bytes20)
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
