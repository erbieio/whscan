package monitor

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"server/database"
	"server/ethhelper"
	"server/log"
	"strconv"
	"time"
)

func SyncBlock() {
	var blockNumber uint64
	currentBlockNumber, err := ethhelper.GetBlockNumber()
	if err != nil {
		log.Infof(err.Error())
	}
	for {
		blockNumber, _ = ethhelper.GetBlockNumber()
		if blockNumber == 0 || blockNumber == currentBlockNumber {
			time.Sleep(5 * time.Second)
			continue
		}
		currentBlockNumber = blockNumber
		var tmp big.Int
		tmp.SetUint64(currentBlockNumber)
		b, err := ethhelper.GetBlock("0x" + hex.EncodeToString(tmp.Bytes()))
		if err != nil {
			log.Infof(err.Error())
		}
		var block database.Block
		hexTo10 := func(v string) string {
			tmp.SetString(v, 0)
			return fmt.Sprintf("%v", tmp.Uint64())
		}
		block.ParentHash = b.UncleHash
		block.UncleHash = b.UncleHash
		block.Coinbase = b.Coinbase
		block.Root = b.Root
		block.TxHash = b.TxHash
		block.ReceiptHash = b.ReceiptHash
		block.Difficulty = b.Difficulty
		block.Number = hexTo10(b.Number)
		block.GasLimit = hexTo10(b.GasLimit)
		block.GasUsed = hexTo10(b.GasUsed)
		block.Extra = b.Extra
		block.MixDigest = b.MixDigest
		block.Size = hexTo10(b.Size)
		block.TxCount = len(b.Transactions)
		block.Ts = hexTo10(b.Ts)
		err = block.Insert()
		if err != nil {
			log.Infof(err.Error())
		}
		type ethTransfer struct {
			From  string `json:"from"`
			To    string `json:"to"`
			Value string `json:"value"`
		}
		var eths []ethTransfer
		for _, t := range b.Transactions {
			var tx database.Transaction
			tx.Hash = t.Hash
			tx.From = t.From
			tx.To = t.To
			tx.Value = hexTo10(t.Value)
			tx.Value = toDecimal(tx.Value, 18)
			tx.Nonce = hexTo10(t.Nonce)
			tx.BlockHash = t.BlockHash
			tx.BlockNumber = hexTo10(t.BlockNumber)
			tx.TransactionIndex = hexTo10(t.TransactionIndex)
			tx.Gas = hexTo10(t.Gas)
			tx.GasPrice = hexTo10(t.GasPrice)
			tx.Input = t.Input

			calls := TraceTxInternalCall(tx.Hash, tx.From, tx.To)
			callsStr, _ := json.Marshal(calls)
			tx.InternalCalls = string(callsStr)
			for _, c := range calls {
				var f float64
				f, _ = strconv.ParseFloat(c.Value, 64)
				if f != 0 {
					eths = append(eths, ethTransfer{
						From:  c.From,
						To:    c.To,
						Value: c.Value,
					})
				}
			}
			ethsStr, _ := json.Marshal(eths)
			tx.InternalValueTransfer = string(ethsStr)
			tx.Ts = block.Ts
			status, ty, tokenTransfers := analysisTokenTransfer(tx.Hash)
			if status == "0x1" {
				tx.Status = "1"
			} else {
				tx.Status = "0"
			}

			tx.TxType = ty
			tokenStr, _ := json.Marshal(tokenTransfers)
			tx.TokenTransfer = string(tokenStr)
			err = tx.Insert()
			if err != nil {
				log.Infof(err.Error())
			}
		}
		blockNumber = currentBlockNumber
	}
}

func toDecimal(src string, decimal int) string {
	var ret string
	if len(src) < decimal {
		ret += "0."
		for i := 0; i < decimal-len(src); i++ {
			ret += "0"
		}
		ret += src
		return ret
	} else {
		preLen := len(src) - decimal
		if preLen == 0 {
			return string(src[0]) + "." + src[1:]
		}
		return src[0:preLen] + "." + src[preLen:]
	}
}
