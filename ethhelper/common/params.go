package common

const (
	//mainPoint = "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"
	MainPoint = "http://192.168.1.235:8561/"
	//MainPoint        = "https://rinkeby.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"
	InfraPoint  = "https://rinkeby.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"
	erc721      = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	erc1155     = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	NFT1155Addr = "0xA1e67a33e090Afe696D7317e05c506d7687Bb2E5"

	IsApproveForAllHash = "0xe985e9c5"

	GetApproved721Hash = "0x081812fc"
	BalanceOf1155Hash  = "0x00fdd58e"
	OwnerOf721Hash     = "0x6352211e"

	Erc721Interface  = "80ac58cd"
	Erc1155Interface = "d9b67a26"

	ERBPayAddress = "0x782936e5A3d929e5536AF48EC85546c4857c70eB"
	CheckAuthHash = "0x4b165090"
)
const (
	WETH          = "0xf4bb2e28688e89fcce3c0580d37d36a7672e8a9f"
	BalanceOfHash = "0x70a08231"
	AllowanceHash = "0xdd62ed3e"
)
const (
	Admin         = "0x56c971ebBC0cD7Ba1f977340140297C0B48b7955"
	AdminListHash = "0x0f560cd7"
	AddIndexHash  = "0x151b01f9"
)

const (
	TradeCore   = "0xD8D5D49182d7Abf3cFc1694F8Ed17742886dDE82"
	AuctionHash = "0x151b01f9"
)
const privKey = "564ea566096d3de340fc5ddac98aef672f916624c8b0e4664a908cd2a2d156fe"
const from = "0x077d34394Ed01b3f31fBd9816cF35d4558146066"

type CallParamTemp struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

type Block struct {
	Transactions []Tx   `json:"transactions" `
	Ts           string `json:"timestamp" `
}
type Tx struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}
type Log struct {
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	Topics           []string `json:"topics"`
	TxHash           string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
}
type Receipt struct {
	TransactionHash   string `json:"transactionHash"`
	TransactionIndex  string `json:"transactionIndex"`
	BlockNumber       string `json:"blockNumber"`
	BlockHash         string `json:"blockHash"`
	Logs              []Log  `json:"logs"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	GasUsed           string `json:"gasUsed"`
	ContractAddress   string `json:"contractAddress"`
	LogsBloom         string `json:"logsBloom"`
	Status            string `json:"status"`
}
type CallParam struct {
	From string `json:"from"`
	To   string `json:"to"`
	Data string `json:"data"`
}

type LogFilter struct {
	FromBlock string   `json:"fromBlock"`
	ToBlock   string   `json:"toBlock"`
	Topics    []string `json:"topics"`
}
type RawData struct {
	Data string `json:"data"`
}
