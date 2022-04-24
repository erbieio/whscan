package service

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
