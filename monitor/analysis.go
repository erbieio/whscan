package monitor

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"server/database"
	"server/ethhelper"
	"strings"
)

type TokenTransfer struct {
	From    string
	To      string
	Value   string
	Name    string
	Symbol  string
	TokenId string
	Type    string
}

func analysisTokenTransfer(txHash string) (string, string, []TokenTransfer) {
	var data []TokenTransfer
	receipts, err := ethhelper.TransactionReceipt(txHash)
	var ty string
	if err != nil {

	}

	for _, log := range receipts.Logs {
		var logDb database.TxLog
		logDb.TxHash = log.TxHash
		logDb.Data = log.Data
		logDb.Address = txHash
		for _, t := range log.Topics {
			logDb.Topics += "," + t
		}
		if len(logDb.Topics) != 0 {
			logDb.Topics = logDb.Topics[1:]
		}
		err = logDb.Insert()
		if err != nil {

		}
	}

	var tmp, from, to, value, tokenId big.Int
	for _, d := range receipts.Logs {
		if d.Topics[0] == erc1155TransferEvent {
			from.SetString(d.Topics[2], 0)
			to.SetString(d.Topics[3], 0)
			dataTmp := d.Data[2:]
			if len(d.Topics) >= 6 {
				tokenId.SetString(d.Topics[4], 0)
				value.SetString(d.Topics[5], 0)
			} else {
				if len(d.Data) > 100 {
					tokenId.SetString(dataTmp[:64], 0)
					value.SetString(dataTmp[64:], 0)
				}
			}

			data = append(data, TokenTransfer{
				From:    common.BytesToAddress(from.Bytes()).String(),
				To:      common.BytesToAddress(to.Bytes()).String(),
				TokenId: tokenId.String(),
				Value:   value.String(),
				Type:    "ERC1155",
			})
			ty += "ERC1155"
		} else if d.Topics[0] == erc721Or20TransferEvent {
			re := ethhelper.ContractCall(d.Address, erc721Input)
			tmp.SetString(re, 0)

			from.SetString(d.Topics[1], 0)
			to.SetString(d.Topics[2], 0)
			value.SetString(d.Data, 0)

			name, _ := ethhelper.CallViewFunc(d.Address, nameHash)
			symbol, _ := ethhelper.CallViewFunc(d.Address, symbolHash)
			name = name[130:]
			name = strings.TrimRight(name, "0")
			symbol = symbol[130:]
			symbol = strings.TrimRight(symbol, "0")
			bytes, _ := hex.DecodeString(name)
			name = string(bytes)
			symbol = symbol[2:]
			bytes, _ = hex.DecodeString(symbol)
			symbol = string(bytes)
			item := TokenTransfer{
				From:   common.BytesToAddress(from.Bytes()).String(),
				To:     common.BytesToAddress(to.Bytes()).String(),
				Name:   name,
				Symbol: symbol,
			}
			//erc721
			if tmp.Uint64() == 1 {
				item.Type = "ERC721"
				item.TokenId = value.String()
				ty += "ERC721"
			} else { //erc20
				item.Type = "ERC20"
				ty += "ERC20"
				item.Value = value.String()
				decimal, _ := ethhelper.Decimals(d.Address)
				item.Value = toDecimal(item.Value, int(decimal))
			}
			data = append(data, item)
		}
	}
	return receipts.Status,ty, data
}
