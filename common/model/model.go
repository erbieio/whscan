// Package model database table definition
package model

import (
	"math/big"

	"gorm.io/gorm"
	"server/common/types"
)

var Tables = []interface{}{
	&Stats{},
	&Block{},
	&Transaction{},
	&EventLog{},
	&Account{},
	&InternalTx{},
	&ERC20Transfer{},
	&ERC721Transfer{},
	&ERC1155Transfer{},
	&Pledge{},
	&Staker{},
	&Slashing{},
	&Validator{},
	&NFT{},
	&Epoch{},
	&Creator{},
	&SNFT{},
	&Erbie{},
	&Reward{},
	&Location{},
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(Tables...)
}

func ClearTable(db *gorm.DB) (err error) {
	if err = db.Migrator().DropTable(Tables...); err == nil {
		err = db.AutoMigrate(Tables...)
	}
	return
}

// Stats caches some database queries to speed up queries
type Stats struct {
	Ready                bool   `json:"ready" gorm:"-"`                        //ready, sync latest block
	ChainId              int64  `json:"chainId" gorm:"primaryKey"`             //chain id
	GenesisBalance       string `json:"genesisBalance" gorm:"type:CHAR(128)"`  //Total amount of coins created
	TotalAmount          string `json:"totalAmount" gorm:"type:CHAR(128)"`     //total transaction volume
	TotalNFTAmount       string `json:"totalNFTAmount" gorm:"type:CHAR(128)"`  //Total transaction volume of NFTs
	TotalSNFTAmount      string `json:"totalSNFTAmount" gorm:"type:CHAR(128)"` //Total transaction volume of SNFTs
	TotalRecycle         int64  `json:"totalRecycle" gorm:"-"`                 //Total number of recycle SNFT
	AvgBlockTime         int64  `json:"avgBlockTime" gorm:"-"`                 //average block time, ms
	TotalBlock           int64  `json:"totalBlock" gorm:"-"`                   //Total number of blocks
	TotalBlackHole       int64  `json:"totalBlackHole" gorm:"-"`               //Total number of BlackHole blocks
	TotalTransaction     int64  `json:"totalTransaction" gorm:"-"`             //Total number of transactions
	TotalInternalTx      int64  `json:"totalInternalTx" gorm:"-"`              //Total number of internal transactions
	TotalTransferTx      int64  `json:"totalTransferTx" gorm:"-"`              //Total number of  transfer transactions
	TotalErbieTx         int64  `json:"totalErbieTx" gorm:"-"`                 //Total number of  erbie transactions
	TotalAccount         int64  `json:"totalAccount" gorm:"-"`                 //Total account number of used
	TotalBalance         string `json:"totalBalance" gorm:"-"`                 //The total amount of coins in the chain
	ActiveAccount        int64  `json:"activeAccount" gorm:"-"`                //The number of active account
	TotalStaker          int64  `json:"totalStaker" gorm:"-"`                  //Total number of stakers
	TotalNFT             int64  `json:"totalNFT" gorm:"-"`                     //Total number of NFTs
	TotalSNFT            int64  `json:"totalSNFT" gorm:"-"`                    //Total number of SNFTs
	TotalNFTTx           int64  `json:"totalNFTTx" gorm:"-"`                   //Total number of  NFT transactions
	TotalSNFTTx          int64  `json:"totalSNFTTx" gorm:"-"`                  //Total number of  SNFT transactions
	TotalValidatorOnline int64  `json:"totalValidatorOnline" gorm:"-"`         //Total amount of validator online
	TotalValidator       int64  `json:"totalValidator" gorm:"-"`               //Total number of validator
	TotalNFTCreator      int64  `json:"totalNFTCreator" gorm:"-"`              //Total creator of NFTs
	TotalSNFTCreator     int64  `json:"totalSNFTCreator" gorm:"-"`             //Total creator of SNFTs
	TotalStakerTx        int64  `json:"totalStakerTx" gorm:"-"`                //Total number of staker  transactions
	RewardCoinCount      int64  `json:"rewardCoinCount" gorm:"-"`              //Total number of times to get coin rewards, 0.1ERB once
	RewardSNFTCount      int64  `json:"rewardSNFTCount" gorm:"-"`              //Total number of times to get SNFT rewards
	TotalPledge          string `json:"totalPledge" gorm:"-"`                  //Total amount of  pledge
	Total24HStakerTx     int64  `json:"total24HStakerTx" gorm:"-"`             //Total number of staker transactions within 24 hours
	Total24HNFT          int64  `json:"total24HNFT" gorm:"-"`                  //Total number of NFT within 24 hours
	Total24HTx           int64  `json:"total24HTx" gorm:"-"`                   //Total number of transactions within 24 hours
	TotalEpoch           int64  `json:"totalEpoch" gorm:"-"`                   //Total number of epoch
	TotalCreator         int64  `json:"totalCreator" gorm:"-"`                 //Total number of creator
	TotalProfit          string `json:"totalProfit" gorm:"-"`                  //Total number of creator profit

	Genesis  Header                     `json:"-" gorm:"-"`
	Balances map[types.Address]*big.Int `json:"-" gorm:"-"`
}

// Header block header information
type Header struct {
	Difficulty       types.Long    `json:"difficulty"`                              //difficulty
	ExtraData        string        `json:"extraData"`                               //Extra data
	GasLimit         types.Long    `json:"gasLimit"`                                //Gas limit
	GasUsed          types.Long    `json:"gasUsed"`                                 //Gas consumption
	Hash             types.Hash    `json:"hash" gorm:"type:CHAR(66);primaryKey"`    //Hash
	Miner            types.Address `json:"miner" gorm:"type:CHAR(42);index"`        //miner
	MixHash          types.Hash    `json:"mixHash" gorm:"type:CHAR(66)"`            //Mixed hash
	Nonce            types.Bytes8  `json:"nonce" gorm:"type:CHAR(18)"`              //difficulty random number
	Number           types.Long    `json:"number" gorm:"index"`                     //block number
	ParentHash       types.Hash    `json:"parentHash" gorm:"type:CHAR(66)"`         //parent block hash
	ReceiptsRoot     types.Hash    `json:"receiptsRoot" gorm:"type:CHAR(66)"`       //Transaction receipt root hash
	Sha3Uncles       types.Hash    `json:"sha3Uncles" gorm:"type:CHAR(66)"`         //Uncle root hash
	Size             types.Long    `json:"size"`                                    //size
	StateRoot        types.Hash    `json:"stateRoot" gorm:"type:CHAR(66)"`          //World tree root hash
	Timestamp        types.Long    `json:"timestamp" gorm:"index"`                  //timestamp
	TotalDifficulty  types.BigInt  `json:"totalDifficulty" gorm:"type:DECIMAL(65)"` //total difficulty
	TransactionsRoot types.Hash    `json:"transactionsRoot" gorm:"type:CHAR(66)"`   //transaction root hash
}

// Block information
type Block struct {
	Header
	Uncles           []types.Hash    `json:"uncles" gorm:"type:VARCHAR(139);serializer:json"` //Uncle block hash
	TotalTransaction types.Long      `json:"totalTransaction"`                                //number of transactions
	Proposers        []types.Address `json:"proposers,omitempty" gorm:"serializer:json"`      //black hole block proposers address
	Proof            []types.Hash    `json:"proof,omitempty" gorm:"index;serializer:json"`    //slash validator proof, multi-signature block hash
}

// Slashing validator and staker penalty record
type Slashing struct {
	Address     types.Address `json:"address" gorm:"type:CHAR(42);index"`       //account address, validator or staker
	BlockNumber types.Long    `json:"block_number" gorm:"index"`                //block number
	Amount      *types.BigInt `json:"amount,omitempty" gorm:"type:DECIMAL(65)"` //penalty amount, unit wei
	Weight      types.Long    `json:"weight,omitempty"`                         //weight after penalty
	Reason      string        `json:"reason" gorm:"type:VARCHAR(42)"`           //penalty reason, 1: no block; 2: multi-signature; address: validator penalty
}

// Account information
type Account struct {
	Address   types.Address       `json:"address" gorm:"type:CHAR(42);primaryKey"`  //address
	Balance   types.BigInt        `json:"balance" gorm:"type:DECIMAL(65);index"`    //The total amount of coins in the chain
	Nonce     types.Long          `json:"nonce"`                                    //transaction random number, transaction volume
	Code      *string             `json:"code"`                                     //bytecode
	Number    types.Long          `json:"number" gorm:"index"`                      //last update block number
	Name      *string             `json:"name,omitempty" gorm:"type:VARCHAR(66)"`   //name
	Symbol    *string             `json:"symbol,omitempty" gorm:"type:VARCHAR(66)"` //symbol
	Type      *types.ContractType `json:"type,omitempty"`                           //contract types, ERC20, ERC721, ERC1155
	Creator   *types.Address      `json:"creator,omitempty" gorm:"type:CHAR(42)"`   //the creator, the contract account has value
	CreatedTx *types.Hash         `json:"createdTx,omitempty" gorm:"type:CHAR(66)"` //create transaction
	SNFTCount int64               `json:"snftCount"`                                //hold SNFT number
	SNFTValue string              `json:"snftValue" gorm:"type:DECIMAL(65)"`        //hold SNFT value
	NFTCount  int64               `json:"nftCount"`                                 //hold NFT number
}

// Transaction information
type Transaction struct {
	BlockHash         types.Hash     `json:"blockHash" gorm:"type:CHAR(66)"`          //Block Hash
	BlockNumber       types.Long     `json:"blockNumber" gorm:"index"`                //block number
	Timestamp         types.Long     `json:"timestamp"`                               //The event stamp of the block it is in
	From              types.Address  `json:"from" gorm:"type:CHAR(42);index"`         //Send address
	To                *types.Address `json:"to" gorm:"type:CHAR(42);index"`           //Receive address
	Input             string         `json:"input" gorm:"type:TEXT"`                  //Additional input data, contract call encoded data
	Value             types.BigInt   `json:"value" gorm:"type:DECIMAL(65)"`           //Amount, unit wei
	Nonce             types.Long     `json:"nonce"`                                   //Random number, the number of transactions initiated by the account
	Gas               types.Long     `json:"gas"`                                     //fuel
	GasPrice          types.Long     `json:"gasPrice"`                                //Gas price
	Hash              types.Hash     `json:"hash" gorm:"type:CHAR(66);primaryKey"`    //Hash
	Status            *types.Long    `json:"status,omitempty"`                        //Status, 1: success; 0: failure
	CumulativeGasUsed types.Long     `json:"cumulativeGasUsed"`                       //Cumulative gas consumption
	ContractAddress   *types.Address `json:"contractAddress" gorm:"type:CHAR(42)"`    //The created contract address
	GasUsed           types.Long     `json:"gasUsed"`                                 //Gas consumption
	TxIndex           types.Long     `json:"transactionIndex"`                        //The serial number in the block
	Error             *string        `json:"error,omitempty" gorm:"type:VARCHAR(66)"` //exec error
}

// EventLog transaction log
type EventLog struct {
	Address     types.Address `json:"address" gorm:"type:CHAR(42)"`                          //The contract address
	Topics      []types.Hash  `json:"topics" gorm:"type:VARCHAR(277);serializer:json"`       //topic
	Data        string        `json:"data"`                                                  //data
	Removed     bool          `json:"removed"`                                               //whether to remove
	BlockNumber types.Long    `json:"blockNumber"`                                           //block number
	TxHash      types.Hash    `json:"transactionHash" gorm:"type:CHAR(66);primaryKey;index"` //The transaction hash
	Index       types.Long    `json:"logIndex" gorm:"primaryKey"`                            //The serial number in the transaction
}

// InternalTx internal transaction
type InternalTx struct {
	TxHash types.Hash    `json:"txHash" gorm:"type:CHAR(66);index"` //The transaction
	Index  types.Long    `json:"index" gorm:"not null"`             //call index
	Op     string        `json:"op" gorm:"type:VARCHAR(16)"`        //Operation
	From   types.Address `json:"from" gorm:"type:CHAR(42);index"`   //Originating address
	To     types.Address `json:"to" gorm:"type:CHAR(42);index"`     //Receive address
	Value  types.BigInt  `json:"value" gorm:"type:DECIMAL(65)"`     //Amount, unit wei
	Gas    types.Long    `json:"gas"`                               //Gas
}

// ERC20Transfer ERC20 contract transfer event
type ERC20Transfer struct {
	TxHash  types.Hash    `json:"txHash" gorm:"type:CHAR(66);index"`  //The transaction hash
	Address types.Address `json:"address" gorm:"type:CHAR(42);index"` //Contract address
	From    types.Address `json:"from" gorm:"type:CHAR(42);index"`    //Originating address
	To      types.Address `json:"to" gorm:"type:CHAR(42);index"`      //Receive address
	Value   types.BigInt  `json:"value" gorm:"type:VARCHAR(80)"`      //amount
}

// ERC721Transfer ERC721 contract transfer event
type ERC721Transfer struct {
	TxHash  types.Hash    `json:"txHash" gorm:"type:CHAR(66);index"` //The transaction hash
	Address types.Address `json:"address" gorm:"type:CHAR(42)"`      //Contract address
	From    types.Address `json:"from" gorm:"type:CHAR(42);index"`   //Originating address
	To      types.Address `json:"to" gorm:"type:CHAR(42);index"`     //Receive address
	TokenId types.BigInt  `json:"tokenId" gorm:"type:VARCHAR(80)"`   //Token ID
}

// ERC1155Transfer ERC1155 contract transfer event
type ERC1155Transfer struct {
	TxHash   types.Hash    `json:"txHash" gorm:"type:CHAR(66);index"` //The transaction hash
	Address  types.Address `json:"address" gorm:"type:CHAR(42)"`      //Contract address
	Operator types.Address `json:"operator" gorm:"type:CHAR(42)"`     //operator address
	From     types.Address `json:"from" gorm:"type:CHAR(42);index"`   //Originating address
	To       types.Address `json:"to" gorm:"type:CHAR(42);index"`     //Receive address
	TokenId  types.BigInt  `json:"tokenId" gorm:"type:VARCHAR(80)"`   //Token ID
	Value    types.BigInt  `json:"value" gorm:"type:VARCHAR(80)"`     //Token amount
}

// NFT User NFT attribute information
type NFT struct {
	Address      string  `json:"address" gorm:"type:CHAR(42);primary_key"` //NFT address, grows automatically from 0x1
	RoyaltyRatio int64   `json:"royalty_ratio"`                            //Royalty rate, in ten thousandths
	MetaUrl      string  `json:"meta_url" gorm:"type:VARCHAR(8192)"`       //Real meta information URL
	LastPrice    *string `json:"last_price" gorm:"type:DECIMAL(65)"`       //The last transaction price (null if the transaction is not completed), the unit is wei
	TxAmount     string  `json:"tx_amount" gorm:"type:DECIMAL(65);index"`  //the total transaction volume of this NFT
	Creator      string  `json:"creator" gorm:"type:CHAR(42)"`             //Creator address
	Timestamp    int64   `json:"timestamp" gorm:"index"`                   //Create timestamp
	BlockNumber  int64   `json:"block_number"`                             //The height of the created block
	TxHash       string  `json:"tx_hash" gorm:"type:CHAR(66)"`             //The transaction hash created
	Owner        string  `json:"owner" gorm:"type:CHAR(42);index"`         //owner
}

// Epoch SNFT Phase 1
// One SNFT->16 Collections->16 FNFTs->256 SNFTs
type Epoch struct {
	ID           string `json:"id" gorm:"type:CHAR(39);primary_key"` //period ID
	Creator      string `json:"creator" gorm:"type:CHAR(42);index"`  //creator address, also the address of royalty income
	RoyaltyRatio int64  `json:"royaltyRatio"`                        //the royalty rate of the same period of SNFT, the unit is one ten thousandth
	MetaUrl      string `json:"meta_url" gorm:"type:VARCHAR(8192)"`  //Real meta information URL
	WeightValue  string `json:"weightValue"`                         //snft value
	WeightAmount int64  `json:"weightAmount"`                        //hold block number amount
	Voter        string `json:"voter" gorm:"type:CHAR(42);index"`    //voter
	Reward       string `json:"reward" gorm:"type:DECIMAL(65)"`      //vote reward amount
	Profit       string `json:"profit" gorm:"type:DECIMAL(65)"`      //royalty profit amount
	Number       int64  `json:"number"`                              //is selected block height
	Timestamp    int64  `json:"timestamp"`                           //is selected timestamp
	StartNumber  int64  `json:"startNumber"`                         //starting the period block height
	StartTime    int64  `json:"startTime"`                           //starting the period timestamp
}

// Creator SNFT creator information
type Creator struct {
	Address    string `json:"address" gorm:"type:CHAR(42);primaryKey"` //account address
	Number     int64  `json:"number"`                                  //be a Creator block number
	Timestamp  int64  `json:"timestamp"`                               //be a Creator Time
	LastEpoch  string `json:"lastEpoch" gorm:"type:CHAR(39)"`          //last selected for the epoch
	LastNumber int64  `json:"lastNumber"`                              //last selected for the block number
	LastTime   int64  `json:"lastTime"`                                //last selected for the timestamp
	Count      int64  `json:"count"`                                   //selected count
	Reward     string `json:"reward" gorm:"type:DECIMAL(65);index"`    //vote profit
	Profit     string `json:"profit" gorm:"type:DECIMAL(65);index"`    //royalty profit
}

// SNFT of SNFT fragments
type SNFT struct {
	Address      string  `json:"address" gorm:"type:VARCHAR(42);primaryKey"` //SNFT address
	LastPrice    *string `json:"last_price" gorm:"type:DECIMAL(65)"`         //The last transaction price, the unit is wei, null if the transaction has not been completed
	TxAmount     string  `json:"tx_amount" gorm:"type:DECIMAL(65);index"`    //the total transaction volume of this SNFT
	RewardAt     int64   `json:"reward_at"`                                  //The timestamp of the last rewarded, null if not rewarded
	RewardNumber int64   `json:"reward_number" gorm:"index"`                 //The height of the last rewarded block
	Owner        string  `json:"owner" gorm:"type:CHAR(42);index"`           //owner, unallocated and reclaimed are null
	Pieces       int64   `json:"pieces"`                                     //snft pieces number
	Remove       bool    `json:"remove" gorm:"index"`                        //SNFTs that are synthesized and then removed
}

// Erbie erbie transaction attribute information
type Erbie struct {
	TxHash      string `json:"tx_hash" gorm:"type:CHAR(66);primary_key"`        //owned transaction hash
	Type        uint8  `json:"type" gorm:"index"`                               //transaction type
	Address     string `json:"address,omitempty" gorm:"type:VARCHAR(42);index"` //the NFT or SNFT address of the transaction
	From        string `json:"from" gorm:"type:CHAR(42);index"`                 //seller or caller address
	To          string `json:"to" gorm:"type:CHAR(42);index"`                   //buyer or proxy address to set
	Value       string `json:"value" gorm:"type:DECIMAL(65)"`                   //price value, the unit is wei
	Extra       string `json:"extra,omitempty" gorm:"type:VARCHAR(8192)"`       //fee receive address or meta url
	Timestamp   int64  `json:"timestamp"`                                       //transaction timestamp
	BlockNumber int64  `json:"block_number" gorm:"index"`                       //block number
	RoyaltyRate int64  `json:"royalty_rate,omitempty"`                          //for the creator royalty rate
	FeeRate     int64  `json:"fee_rate,omitempty"`                              //fee rate, unit wei; or number of recycle snft pieces
	Proxy       string `json:"proxy,omitempty"`                                 //It's validator's proxy address to run a node
}

// Reward miner reward, the reward method is SNFT and Amount, and Amount is tentatively set to 0.1ERB
type Reward struct {
	Address     string  `json:"address" gorm:"type:CHAR(42);index"`       //reward address
	Proxy       *string `json:"proxy,omitempty" gorm:"type:CHAR(42)"`     //proxy address
	Identity    uint8   `json:"identity"`                                 //Identity, 1: block producer, 2: validator, 3, staker
	BlockNumber int64   `json:"block_number" gorm:"index"`                //The block number when rewarding
	SNFT        string  `json:"snft,omitempty" gorm:"type:CHAR(42)"`      //SNFT address
	Amount      *string `json:"amount,omitempty" gorm:"type:DECIMAL(65)"` //Amount of reward
}

// Pledge records from stakers to validators
type Pledge struct {
	Staker      string `json:"staker" gorm:"type:CHAR(42);primary_key;index"`    //staker address
	Validator   string `json:"validator" gorm:"type:CHAR(42);primary_key;index"` //validator address
	Amount      string `json:"amount" gorm:"type:DECIMAL(65)"`                   //pledge amount
	Timestamp   int64  `json:"timestamp" gorm:"index"`                           //latest time
	BlockNumber int64  `json:"block_number" gorm:"index"`                        //latest block
	TxHash      string `json:"tx_hash" gorm:"type:CHAR(66)"`                     //the transaction created
}

// Staker staker attribute information
type Staker struct {
	Address     string `json:"address" gorm:"type:CHAR(42);primary_key"` //staker address
	Timestamp   int64  `json:"timestamp" gorm:"index"`                   //create time
	BlockNumber int64  `json:"block_number" gorm:"index"`                //the block number when created
	TxHash      string `json:"tx_hash" gorm:"type:CHAR(66)"`             //the record created
	Amount      string `json:"amount" gorm:"type:DECIMAL(65)"`           //pledge amount
	FeeRate     int64  `json:"fee_rate"`                                 //fee charged when nft or snft transaction
	Reward      string `json:"reward" gorm:"type:DECIMAL(65);index"`     //amount of total reward
	RewardCount int64  `json:"reward_count"`                             //reward snft count
}

// Validator pledge information
type Validator struct {
	Address      string `json:"address" gorm:"type:CHAR(42);primaryKey"` //staking account
	Proxy        string `json:"proxy" gorm:"type:CHAR(42)"`              //proxy address
	Amount       string `json:"amount" gorm:"type:DECIMAL(65)"`          //pledge amount
	Reward       string `json:"reward" gorm:"type:DECIMAL(65)"`          //amount of total reward
	RewardCount  int64  `json:"reward_count"`                            //reward coin count
	RewardNumber int64  `json:"reward_number"`                           //The number at latest reward
	Timestamp    int64  `json:"timestamp"`                               //The time at latest updated
	BlockNumber  int64  `json:"block_number"`                            //The block number at latest updated
	TxHash       string `json:"tx_hash" gorm:"type:CHAR(66)"`            //the record created
	Weight       int64  `json:"weight"`                                  //online weight,if it is not 70, it means that it is not online
	Score        int64  `json:"score"`                                   //node comprehensive score
}

type Location struct {
	Address   string  `json:"address" gorm:"type:CHAR(42);primaryKey"` //account address
	IP        string  `json:"ip" gorm:"type:VARCHAR(15)"`              //account ip
	Latitude  float64 `json:"latitude"`                                //latitude
	Longitude float64 `json:"longitude"`                               //longitude
	City      string  `json:"city"`                                    //city
	Country   string  `json:"country"`                                 //country
}

// Parsed block parsing result
type Parsed struct {
	*Block
	CacheTxs          []*Transaction `json:"transactions"`
	CacheInternalTxs  []*InternalTx
	CacheTransferLogs []interface{}
	CacheAccounts     []*Account
	CacheLogs         []*EventLog

	// erbie, which need to be inserted into the database by priority (later data may query previous data)
	Epoch     *Epoch      //Official injection of the first phase of SNFT
	Slashings []*Slashing //penalty slash
	Erbies    []*Erbie    //NFT transaction record, transfer, recycle, pledge
	Rewards   []*Reward   //reward record
	Mergers   []*SNFT     //merge snft
}
