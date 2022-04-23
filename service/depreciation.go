package service

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"server/conf"
	"server/ethclient"
)

type Block struct {
	Number           string `json:"number"`
	ParentHash       string `json:"parentHash"`
	Sha3Uncles       string `json:"sha3Uncles"`
	Miner            string `json:"miner"`
	Timestamp        string `json:"timestamp" `
	TotalTransaction int    `json:"tx_count" gorm:"column:total_transaction"`
	StateRoot        string `json:"stateRoot"`
	TransactionsRoot string `json:"transactionsRoot"`
	ReceiptsRoot     string `json:"receiptsRoot"`
	GasLimit         string `json:"gasLimit"`
	GasUsed          string `json:"gasUsed"`
	Difficulty       string `json:"difficulty"`
	Size             string `json:"size"`
	MixHash          string `json:"mixHash"`
	ExtraData        string `json:"extraData"`
}

func FetchBlocks_(page, size int) (data []Block, count int64, err error) {
	err = DB.Model(&Block{}).Count(&count).Error
	err = DB.Model(&Block{}).Limit(size).Offset((page - 1) * size).Order("number DESC").Find(&data).Error
	return data, count, err
}

func FindBlock(num string) (data Block, err error) {
	db := DB.Model(&Block{}).Where("number=?", num)
	err = db.First(&data).Error
	return data, err
}

type Transaction struct {
	Hash                  string   `json:"hash"`
	BlockHash             string   `json:"blockHash"`
	BlockNumber           string   `json:"blockNumber"`
	Timestamp             string   `json:"timestamp"`
	From                  string   `json:"from"`
	To                    string   `json:"to"`
	Value                 string   `json:"value"`
	Gas                   string   `json:"gas"`
	GasPrice              string   `json:"gasPrice"`
	TxIndex               string   `json:"transactionIndex"`
	Nonce                 string   `json:"nonce"`
	Input                 string   `json:"input"`
	Status                string   `json:"status"`
	InternalValueTransfer StrArray `json:"internal_value_transfer"`
	InternalCalls         StrArray `json:"internal_calls"`
	TokenTransfer         StrArray `json:"token_transfer"`
}

type StrArray string

func (t StrArray) MarshalJSON() ([]byte, error) {
	return []byte("\"[]\""), nil
}

func FetchTxs(page, size int, addr, block string) (data []Transaction, count int64, err error) {
	db := DB.Model(&Transaction{})
	if addr != "" {
		db = db.Where("`from`=? or `to`=?", addr, addr)
	}
	if block != "" {
		db = db.Where("block_number=?", block)
	}
	err = db.Count(&count).Error
	err = db.Limit(size).Offset((page - 1) * size).Joins("LEFT JOIN blocks ON transactions.block_hash=blocks.hash").
		Select("transactions.*, blocks.timestamp AS timestamp").Order("block_number DESC, tx_index DESC").Scan(&data).Error
	return data, count, err
}

func FindTx(txHash string) (data Transaction, err error) {
	err = DB.Model(Transaction{}).Where("transactions.hash=?", txHash).Joins("LEFT JOIN blocks ON transactions.block_hash=blocks.hash").
		Select("transactions.*, blocks.timestamp AS timestamp").First(&data).Error
	return data, err
}

type Log struct {
	Address string `json:"address"`
	Topics  string `json:"topics"`
	Data    string `json:"data"`
	TxHash  string `json:"tx_hash"`
}

func FindLogByTx(txHash string) (data []Log, err error) {
	err = DB.Model(&Log{}).Where("tx_hash=?", txHash).Find(&data).Error
	return data, err
}

func ExchangerCheck(addr string) (bool, string, uint64, error) {
	client, err := ethclient.Dial(conf.ChainUrl)
	if err != nil {
		return false, "", 0, err
	}
	exchanger, err := client.GetExchanger(addr)
	if err != nil {
		return false, "", 0, err
	}
	status, err := checkAuth(client, addr)
	return exchanger.ExchangerFlag, exchanger.ExchangerBalance.String(), status, err
}

type CallParamTemp struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

func checkAuth(client *ethclient.Client, addr string) (uint64, error) {
	var tmp big.Int
	payload := make([]byte, 36)

	tmp.SetString("0x4b165090", 0)
	copy(payload[:4], tmp.Bytes())
	tmp.SetString(addr, 0)
	copy(payload[36-len(tmp.Bytes()):], tmp.Bytes())
	params := CallParamTemp{To: conf.ERBPay, Data: "0x" + hex.EncodeToString(payload)}
	jsonData, err := json.Marshal(params)
	if err != nil {
		return 0, fmt.Errorf("Umarshal failed:" + err.Error() + string(jsonData))
	}
	var ret string
	if err = client.Call(&ret, "eth_call", params, "latest"); err != nil {
		return 0, fmt.Errorf("Call failed:" + err.Error())
	} else {
		tmp.SetString(ret, 0)
		return tmp.Uint64(), nil
	}
}
