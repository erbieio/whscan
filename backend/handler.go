package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"server/common/model"
	"server/common/types"
	"server/common/utils"
	"server/node"
	"server/service"
)

var NotFound = fmt.Errorf("not found")

// decode parses the block
func decode(c *node.Client, ctx context.Context, number types.Long) (parsed *model.Parsed, err error) {
	// Get the block (including the transaction)
	err = c.CallContext(ctx, &parsed, "eth_getBlockByNumber", number.Hex(), true)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	} else if parsed == nil {
		return nil, NotFound
	}
	// Get transaction receipt (including transaction log)
	parsed.TotalTransaction = types.Long(len(parsed.CacheTxs))
	if parsed.TotalTransaction > 0 {
		// get transaction receipt
		reqs := make([]node.BatchElem, parsed.TotalTransaction)
		for i := range reqs {
			reqs[i] = node.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []any{parsed.CacheTxs[i].Hash},
				Result: &parsed.CacheTxs[i],
			}
		}
		if err = c.BatchCallContext(ctx, reqs); err != nil {
			return nil, fmt.Errorf("eth_getTransactionReceipt err:%v", err)
		}
		for i := range reqs {
			if reqs[i].Error != nil || parsed.CacheTxs[i].GasUsed == 0 {
				return nil, fmt.Errorf("eth_getTransactionReceipt receipt:%v,err:%v", reqs[i].Result, reqs[i].Error)
			}
			parsed.CacheTxs[i].Timestamp = parsed.Timestamp
		}
		// Get the receipt logs, which can only be checked according to the block hash (there may be multiple blocks with the same block height)
		err = c.CallContext(ctx, &parsed.CacheLogs, "eth_getLogs", map[string]any{"blockHash": parsed.Hash})
		if err != nil {
			return nil, fmt.Errorf("eth_getLogs err:%v", err)
		}
	}
	for _, log := range parsed.CacheLogs {
		if transferLog := utils.UnpackTransferLog(log); transferLog != nil {
			parsed.CacheTransferLogs = append(parsed.CacheTransferLogs, transferLog...)
		}
	}
	// Parse changed account properties and internal transactions
	if err = decodeInternalTxs(c, ctx, parsed); err != nil {
		return nil, fmt.Errorf("decodeInternalTxs err:%v", err)
	}
	if err = decodeAccounts(c, ctx, parsed); err != nil {
		return nil, fmt.Errorf("decodeAccounts err:%v", err)
	}
	// Parse things specific to erbie
	err = decodeWH(c, parsed)
	return
}

const tracerCode = `
{
	details: [],
	enter: function (log) {
		const detail = {index:this.details.length, op:log.getType(), from:toHex(log.getFrom()), to:toHex(log.getTo()), gas:log.getGas()};
		detail.value = detail.op==='DELEGATECALL' ? '0x0' : '0x'+log.getValue().toString(16);
		this.details.push(detail);
	},
	exit: function() {},
	step: function() {},
	fault: function() {},
	result: function(ctx) {
		if (this.details.length >= 256) {return {error:ctx.error};}
		for (var i = 0; i < this.details.length; i++) {
			this.details[i].txHash = toHex(ctx.txHash);
		}
		return {details:this.details, error:ctx.error};
	}
}`

// decodeInternalTxs 获取交易的内部调用详细情况
func decodeInternalTxs(c *node.Client, ctx context.Context, parsed *model.Parsed) (err error) {
	for _, tx := range parsed.CacheTxs {
		execRet := struct {
			Details []*model.InternalTx `json:"details"`
			Error   *string             `json:"error"`
		}{}
		if err = c.CallContext(ctx, &execRet, "debug_traceTransaction", tx.Hash, map[string]any{"tracer": tracerCode}); err != nil {
			return
		}
		tx.Error = execRet.Error
		parsed.CacheInternalTxs = append(parsed.CacheInternalTxs, execRet.Details...)
	}
	return
}

// decodeAccount to get account related properties
func decodeAccounts(c *node.Client, ctx context.Context, parsed *model.Parsed) (err error) {
	if parsed.Number == 0 {
		state := struct {
			Accounts map[types.Address]*model.Account `json:"accounts"`
			Next     *string                          `json:"next"`
		}{}
		for next := new(string); next != nil; next, state.Next = state.Next, nil {
			if err = c.CallContext(ctx, &state, "debug_accountRange", "0x0", next, nil, false, true, true); err != nil {
				return
			}
		}
		parsed.CacheAccounts = make([]*model.Account, 0, len(state.Accounts))
		for address, account := range state.Accounts {
			account.Address = address
			account.SNFTValue = "0"
			account.Timestamp = types.Long(time.Now().Local().Unix())
			if account.Code != nil {
				if err = utils.SetProperty(c, ctx, "0x0", account); err != nil {
					return
				}
			}
			parsed.CacheAccounts = append(parsed.CacheAccounts, account)
		}
		return
	}
	// Get the change account address
	var modifiedAccounts []types.Address
	if err = c.CallContext(ctx, &modifiedAccounts, "debug_getModifiedAccountsByHash", parsed.Hash); err != nil {
		return
	}
	if len(modifiedAccounts) > 0 {
		parsed.CacheAccounts = make([]*model.Account, 0, len(modifiedAccounts))
		contracts := make(map[types.Address]*model.Account)
		for _, tx := range parsed.CacheTxs {
			if address := tx.ContractAddress; address != nil {
				contracts[*address] = &model.Account{Creator: &tx.From, CreatedTx: &tx.Hash}
			}
		}
		for _, tx := range parsed.CacheInternalTxs {
			if tx.Op == "CREATE" || tx.Op == "CREATE2" {
				contracts[tx.To] = &model.Account{Creator: &tx.From, CreatedTx: &tx.TxHash}
			}
		}
		number := parsed.Number.Hex()
		info := struct {
			Nonce   types.Long `json:"Nonce"`
			Balance *big.Int   `json:"Balance"`
			//Worm    *struct {
			//	VoteWeight *big.Int `json:"VoteWeight"`
			//} `json:"Worm"`
		}{}
		for _, address := range modifiedAccounts {
			if address != types.ZeroAddress && (address[:12] == "0x0000000000" || address[:12] == "0x8000000000") {
				continue
			}
			account := &model.Account{Address: address, Number: parsed.Number, SNFTValue: "0"}
			if contract := contracts[address]; contract != nil {
				account.Creator, account.CreatedTx = contract.Creator, contract.CreatedTx
				if err = utils.SetProperty(c, ctx, number, account); err != nil {
					return
				}
				if err = c.CallContext(ctx, &account.Code, "eth_getCode", address, number); err != nil {
					return
				}

			}
			if err = c.CallContext(ctx, &info, "eth_getAccountInfo", address, number); err != nil {
				return
			}
			account.Timestamp = types.Long(time.Now().Local().Unix())
			account.Number = parsed.Number
			account.Nonce = info.Nonce
			account.Balance = types.BigInt(info.Balance.String())
			//if info.Worm != nil {
			//	account.SNFTValue = info.Worm.VoteWeight.String()
			//} else {
			//	account.SNFTValue = "0"
			//}
			account.SNFTValue = "0"
			parsed.CacheAccounts = append(parsed.CacheAccounts, account)
		}
	}
	return
}

func write(c *node.Client, ctx context.Context, parsed *model.Parsed) (head types.Long, err error) {
	head, err = service.Insert(parsed)
	if err != nil || parsed.Number == head {
		return
	}
	for parsed.Number = parsed.Number - 2; parsed.Number >= 0; parsed.Number-- {
		number, pass := parsed.Number.Hex(), false
		if err = c.CallContext(ctx, &parsed.Block, "eth_getBlockByNumber", number, false); err != nil {
			return
		}
		if pass, err = service.VerifyHead(parsed); err != nil {
			return
		} else if pass {
			for _, account := range parsed.CacheAccounts {
				info := struct {
					Nonce   types.Long `json:"Nonce"`
					Balance *big.Int   `json:"Balance"`
					//Worm    *struct {
					//	VoteWeight *big.Int `json:"VoteWeight"`
					//} `json:"Worm"`
				}{}
				if err = c.Call(&info, "eth_getAccountInfo", account.Address, number); err != nil {
					return
				}
				account.Number = parsed.Number
				account.Nonce = info.Nonce
				account.Balance = types.BigInt(info.Balance.String())
				//if info.Worm != nil {
				//	account.SNFTValue = info.Worm.VoteWeight.String()
				//} else {
				//	account.SNFTValue = "0"
				//}
				account.SNFTValue = "0"
			}
			break
		}
	}
	return parsed.Number, service.SetHead(parsed)
}

func check(c *node.Client, ctx context.Context) (stats *model.Stats, err error) {
	if err = c.CallContext(ctx, &struct{}{}, "debug_gcStats"); err != nil {
		return
	}
	if err = c.CallContext(ctx, &struct{}{}, "eth_getAccountInfo", types.ZeroAddress, "0x0"); err != nil {
		return
	}
	chainId, genesis, stats := types.Long(0), model.Header{}, service.GetStats()
	if err = c.CallContext(ctx, &chainId, "eth_chainId"); err != nil {
		return
	}
	if err = c.CallContext(ctx, &genesis, "eth_getBlockByNumber", "0x0", false); err != nil {
		return
	}
	if stats.TotalBlock > 0 && (stats.ChainId != int64(chainId) || stats.Genesis != genesis) {
		err = errors.New("stored data and chain node information do not match")
	} else {
		stats.ChainId = int64(chainId)
		stats.Genesis = genesis
	}
	return
}

// decodeWH Imports the underlying NFT transaction of the SNFT meta information distributed by the block
func decodeWH(c *node.Client, wh *model.Parsed) (err error) {
	if number := wh.Number.Hex(); wh.Number > 0 {
		// Miner reward SNFT processing
		//var rewards []*struct {
		//	Address      string   `json:"Address"`
		//	NFTAddress   string   `json:"NftAddress"`
		//	RewardAmount *big.Int `json:"RewardAmount"`
		//}
		type StakerReward struct {
			StakerAddr   string   `json:"StakerAddr"`
			RewardAmount *big.Int `json:"RewardAmount"`
		}
		var rewards []*struct {
			Address       string          `json:"Address"`
			NFTAddress    string          `json:"NftAddress"`
			RewardAmount  *big.Int        `json:"RewardAmount"`
			StakerRewards []*StakerReward `json:"StakerRewards"`
		}
		//err = c.Call(&rewards, "eth_getBlockBeneficiaryAddressByNumber", number, true)
		err = c.Call(&rewards, "eth_getBlockBeneficiaryAddressByNumberNew", number, true)
		if err != nil {
			return fmt.Errorf("GetReward() err:%v", err)
		}
		for i := range rewards {
			identity := uint8(0)
			if wh.Number == 1 {
				identity = 3
			} else {
				switch i {
				case 7, 8, 9, 10:
					identity = 3
				case 0, 1, 2, 3, 4, 5:
					identity = 2
				case 6:
					identity = 1
				default:
					return fmt.Errorf("reward length more than 11")
				}
			}

			validatorReward := &model.Reward{
				Address:     rewards[i].Address,
				Validator:   rewards[i].Address,
				Identity:    identity,
				BlockNumber: int64(wh.Number),
			}
			if rewards[i].RewardAmount == nil {
				// Note that when NFTAddress is zero address error
				validatorReward.SNFT = rewards[i].NFTAddress
			} else {
				validatorReward.Amount = new(string)
				*validatorReward.Amount = rewards[i].RewardAmount.Text(10)
			}
			wh.Rewards = append(wh.Rewards, validatorReward)

			if rewards[i].StakerRewards != nil {
				for _, stakerRd := range rewards[i].StakerRewards {
					stakerReward := &model.Reward{
						Address:     stakerRd.StakerAddr,
						Validator:   rewards[i].Address,
						Identity:    identity,
						BlockNumber: int64(wh.Number),
					}

					stakerReward.Amount = new(string)
					*stakerReward.Amount = stakerRd.RewardAmount.Text(10)
					wh.Rewards = append(wh.Rewards, stakerReward)
				}
			}
		}

		var punished struct {
			PunishedHash       []types.Hash    `json:"punishedHash"`
			PunishedValidators []types.Address `json:"punishedValidators"`
		}
		if err = c.Call(&punished, "eth_getPunishedInfo", number); err != nil {
			return fmt.Errorf("getPunishedInfo() err:%v", err)
		}
		wh.Proof = punished.PunishedHash
		for _, validator := range punished.PunishedValidators {
			wh.Slashings = append(wh.Slashings, &model.Slashing{
				Address:     validator,
				BlockNumber: wh.Number,
				Reason:      "2",
			})
		}

		if wh.Miner == types.ZeroAddress {
			var onlineWeight []*struct {
				Address types.Address `json:"address"`
				Value   types.Long    `json:"value"`
			}
			if err = c.Call(&onlineWeight, "eth_getValidators", number); err != nil {
				return fmt.Errorf("getWeights() err:%v", err)
			}
			for _, weight := range onlineWeight {
				wh.Slashings = append(wh.Slashings, &model.Slashing{
					Address:     weight.Address,
					BlockNumber: wh.Number,
					Weight:      weight.Value,
					Reason:      "1",
				})
			}
			var proposers []*struct {
				Address types.Address
			}
			if err = c.Call(&proposers, "eth_getRealParticipantsByNumber", number); err != nil {
				return fmt.Errorf("getProposers() err:%v", err)
			}
			for _, proposer := range proposers {
				wh.Proposers = append(wh.Proposers, proposer.Address)
			}
		}

		// erbie transaction processing
		for _, tx := range wh.CacheTxs {
			err = decodeWHTx(c, wh, tx)
			if err != nil {
				return err
			}
		}

		// erbie auto merge snft
		//for _, eventLog := range wh.CacheLogs {
		//	if len(eventLog.Topics) == 3 && len(eventLog.Data) >= 66 {
		//		if eventLog.Topics[0] == "0x77415a68a0d28daf11e1308e53371f573e0920810c9cd9de7904777d5fb9d625" {
		//			pieces, _ := strconv.ParseInt(eventLog.Data[62:66], 16, 32)
		//			if pieces > 0 {
		//				addr := string(eventLog.Topics[1][27:])
		//				for i := 0; i < 3; i++ {
		//					if addr[i] == '8' {
		//						wh.Mergers = append(wh.Mergers, &model.SNFT{
		//							Address:      "0x" + addr[i:],
		//							TxAmount:     "0",
		//							RewardAt:     int64(wh.Timestamp),
		//							RewardNumber: int64(wh.Number),
		//							Owner:        "0x" + string(eventLog.Topics[2][26:]),
		//							Pieces:       pieces,
		//						})
		//					}
		//				}
		//			}
		//		}
		//	}
		//}
		//for _, reward := range wh.Rewards {
		//	if reward.SNFT != "" && reward.SNFT[41] == 'f' {
		//		addr := (reward.SNFT)[:41] + "0"
		//		for i := 0; i < 3; i++ {
		//			info := struct {
		//				NFT *struct {
		//					MergeLevel  int    `json:"MergeLevel"`
		//					MergeNumber int64  `json:"MergeNumber"`
		//					Owner       string `json:"Owner"`
		//				} `json:"Nft"`
		//			}{}
		//			if err = c.Call(&info, "eth_getAccountInfo", addr, number); err != nil {
		//				return
		//			}
		//			if info.NFT != nil && info.NFT.MergeLevel > i {
		//				wh.Mergers = append(wh.Mergers, &model.SNFT{
		//					Address:      addr[:41-i],
		//					TxAmount:     "0",
		//					RewardAt:     int64(wh.Timestamp),
		//					RewardNumber: int64(wh.Number),
		//					Owner:        info.NFT.Owner,
		//					Pieces:       info.NFT.MergeNumber,
		//				})
		//				addr = addr[:40-i] + "0" + addr[41+i:]
		//			} else {
		//				break
		//			}
		//		}
		//	}
		//}
	} else {
		info := struct {
			Worm *struct {
				//FeeRate         int64 `json:"FeeRate"`
				ValidatorProxy  string `json:"ValidatorProxy"`
				StakerExtension *struct {
					StakerExtensions []struct {
						Addr    string   `json:"Addr"`
						Balance *big.Int `json:"Balance"`
					} `json:"StakerExtensions"`
				}
			} `json:"Worm"`
		}{}
		for _, account := range wh.CacheAccounts {
			if err = c.Call(&info, "eth_getAccountInfo", account.Address, "0x0"); err != nil {
				return
			}
			if info.Worm != nil && info.Worm.StakerExtension != nil {
				for _, pledge := range info.Worm.StakerExtension.StakerExtensions {
					tempEribie := &model.Erbie{
						TxHash:    "0x0",
						Type:      3,
						From:      string(account.Address),
						To:        pledge.Addr,
						Value:     pledge.Balance.Text(10),
						Timestamp: int64(wh.Timestamp),
						//FeeRate:   info.Worm.FeeRate,
					}
					if string(account.Address) == pledge.Addr {
						if info.Worm.ValidatorProxy != "0x0000000000000000000000000000000000000000" {
							tempEribie.Proxy = info.Worm.ValidatorProxy
						}
					}
					wh.Erbies = append(wh.Erbies, tempEribie)
				}
			}
		}
		//result := struct {
		//	Validators []*struct {
		//		Addr  string `json:"Addr"`
		//		Proxy string `json:"Proxy"`
		//	} `json:"Validators"`
		//}{}
		//if err = c.Call(&result, "eth_getValidator", "0x0"); err != nil {
		//	return
		//}
		//for _, validator := range result.Validators {
		//	wh.Erbies = append(wh.Erbies, &model.Erbie{
		//		TxHash:    "0x0",
		//		Type:      31,
		//		From:      validator.Addr,
		//		To:        validator.Proxy,
		//		Timestamp: int64(wh.Timestamp),
		//	})
		//
		//}
	}
	return
}

// decodeWHTx parses the special transaction of the erbie blockchain
func decodeWHTx(c *node.Client, wh *model.Parsed, tx *model.Transaction) (err error) {
	input, _ := hex.DecodeString(tx.Input[2:])
	// Non-erbie and failed transactions are not resolved
	if *tx.Status == 0 || len(input) < 6 || string(input[0:6]) != "erbie:" {
		return
	}
	w := struct {
		Type         uint8  `json:"type"`
		CSBTAddress  string `json:"csbt_address,omitempty"`
		ProxyAddress string `json:"proxy_address,omitempty"`
		ProxySign    string `json:"proxy_sign,omitempty"`
		Creator      string `json:"creator,omitempty"`
		Version      string `json:"version,omitempty"`
	}{}
	if err = json.Unmarshal(input[6:], &w); err != nil {
		return
	}

	erbie := &model.Erbie{
		TxHash:      string(tx.Hash),
		Type:        w.Type,
		Address:     strings.ToLower(w.CSBTAddress),
		From:        string(tx.From),
		To:          string(*tx.To),
		Value:       string(tx.Value),
		Extra:       "",
		Timestamp:   int64(wh.Timestamp),
		BlockNumber: int64(wh.Number),
		RoyaltyRate: 0,
		FeeRate:     0,
	}
	switch w.Type {
	case 1: //transfer csbt

	case 2: //withdraw erb to owner of csbt

	case 3: //to be a validator or staker
		if tx.From == *tx.To {
			if w.ProxyAddress != "" && w.ProxyAddress != "0x0000000000000000000000000000000000000000" {
				erbie.Proxy = w.ProxyAddress
			}
		}

	case 4: //not to be a validator or staker
		if tx.From == *tx.To {
			info := struct {
				Worm *struct {
					ValidatorProxy     string `json:"ValidatorProxy"`
					ValidatorExtension *struct {
						ValidatorExtensions []struct {
							Addr    string   `json:"Addr"`
							Balance *big.Int `json:"Balance"`
						} `json:"ValidatorExtensions"`
					}
				} `json:"Worm"`
			}{}
			if err = c.Call(&info, "eth_getAccountInfo", tx.From, wh.Number.Hex()); err != nil {
				return
			}

			//获得validator质押金额
			var validatorPledge *big.Int
			for _, validator := range info.Worm.ValidatorExtension.ValidatorExtensions {
				if validator.Addr == string(tx.From) {
					validatorPledge = validator.Balance
				}
			}
			if validatorPledge.Text(10) == string(tx.Value) {
				// 质押在当前validator上的staker都需要一起撤销
				for _, validator := range info.Worm.ValidatorExtension.ValidatorExtensions {
					if validator.Addr == string(tx.From) {
						continue
					}
					erbieStaker := &model.Erbie{
						TxHash:      "0x0",
						Type:        w.Type,
						Address:     strings.ToLower(w.CSBTAddress),
						From:        validator.Addr,
						To:          string(tx.From),
						Value:       validator.Balance.Text(10),
						Extra:       "",
						Timestamp:   int64(wh.Timestamp),
						BlockNumber: int64(wh.Number),
						RoyaltyRate: 0,
						FeeRate:     0,
					}
					wh.Erbies = append(wh.Erbies, erbieStaker)
				}
			}
		}

	case 5: //recover validator online weight

	}
	wh.Erbies = append(wh.Erbies, erbie)
	return
}
