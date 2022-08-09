// Package model database table definition
package model

import (
	"gorm.io/gorm"
	"server/common/types"
)

var Tables = []interface{}{
	&Cache{},
	&Block{},
	&Uncle{},
	&Transaction{},
	&Log{},
	&Account{},
	&InternalTx{},
	&ERC20Transfer{},
	&ERC721Transfer{},
	&ERC1155Transfer{},
	&Exchanger{},
	&NFT{},
	&Epoch{},
	&FNFT{},
	&SNFT{},
	&Collection{},
	&NFTTx{},
	&Reward{},
	&ConsensusPledge{},
	&ExchangerPledge{},
	&Subscription{},
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(Tables...)
}

func DropTable(db *gorm.DB) error {
	return db.Migrator().DropTable(Tables...)
}

var Views = map[string]string{
	"v_snfts": `
CREATE OR REPLACE VIEW vsnfts
AS
SELECT snfts.*, fnfts.*, epoches.creator, epoches.exchanger, royalty_ratio
     , dir, vote_weight, number, collections.id AS collection_id
     , collections.name AS collection_name, collections.desc AS collection_desc
FROM snfts
         LEFT JOIN fnfts ON LEFT(address, 40) = fnfts.id
         LEFT JOIN collections ON LEFT(address, 39) = collections.id
         LEFT JOIN epoches ON LEFT(address, 38) = epoches.id;`,
}

func SetView(db *gorm.DB) error {
	for _, view := range Views {
		err := db.Exec(view).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// Cache stores some statistical data
type Cache struct {
	Key   string `gorm:"type:VARCHAR(64);primaryKey"` //Key
	Value string `gorm:"type:VARCHAR(128)"`           //value
}

// Header block header information
type Header struct {
	Difficulty       types.Uint64   `json:"difficulty"`                            //difficulty
	ExtraData        string         `json:"extraData"`                             //Extra data
	GasLimit         types.Uint64   `json:"gasLimit"`                              //Gas limit
	GasUsed          types.Uint64   `json:"gasUsed"`                               //Gas consumption
	Hash             types.Hash     `json:"hash" gorm:"type:CHAR(66);primaryKey"`  //Hash
	Miner            types.Address  `json:"miner" gorm:"type:CHAR(42)"`            //miner
	MixHash          types.Hash     `json:"mixHash" gorm:"type:CHAR(66)"`          //Mixed hash
	Nonce            types.Data8    `json:"nonce" gorm:"type:CHAR(18)"`            //difficulty random number
	ParentHash       types.Hash     `json:"parentHash" gorm:"type:CHAR(66)"`       //parent block hash
	ReceiptsRoot     types.Hash     `json:"receiptsRoot" gorm:"type:CHAR(66)"`     //Transaction receipt root hash
	Sha3Uncles       types.Hash     `json:"sha3Uncles" gorm:"type:CHAR(66)"`       //Uncle root hash
	Size             types.Uint64   `json:"size"`                                  //size
	StateRoot        types.Hash     `json:"stateRoot" gorm:"type:CHAR(66)"`        //World tree root hash
	Timestamp        types.Uint64   `json:"timestamp"`                             //timestamp
	TransactionsRoot types.Hash     `json:"transactionsRoot" gorm:"type:CHAR(66)"` //transaction root hash
	UncleHashes      types.StrArray `json:"uncles" gorm:"type:TEXT"`               //Uncle block hash
	UnclesCount      types.Uint64   `json:"unclesCount"`                           //Number of uncle blocks
}

// Uncle block information
type Uncle struct {
	Header
	Number types.Uint64 `json:"number"` //Uncle block number
}

// Block information
type Block struct {
	Header
	Number           types.Uint64 `json:"number" gorm:"uniqueIndex"`               //block number
	TotalDifficulty  types.BigInt `json:"totalDifficulty" gorm:"type:VARCHAR(64)"` //Total difficulty
	TotalTransaction types.Uint64 `json:"totalTransaction"`                        //Number of transactions
}

// Account information
type Account struct {
	Address   types.Address  `json:"address" gorm:"type:CHAR(42);primaryKey"` //address
	Balance   types.BigInt   `json:"balance" gorm:"type:VARCHAR(128);index"`  //balance
	Nonce     types.Uint64   `json:"transactionCount"`                        //Transaction random number, transaction volume
	Code      *string        `json:"code"`                                    //bytecode
	Name      *string        `json:"name" gorm:"type:VARCHAR(64)"`            //name
	Symbol    *string        `json:"symbol" gorm:"type:VARCHAR(64)"`          //symbol
	ERC       types.ERC      `json:"ERC"`                                     //ERC types, ERC20, ERC721, ERC1155
	Creator   *types.Address `json:"creator" gorm:"type:CHAR(42)"`            //The creator, the contract account has value
	CreatedTx *types.Hash    `json:"createdTx" gorm:"type:CHAR(66)"`          //Create transaction
}

// Transaction information
type Transaction struct {
	BlockHash   types.Hash     `json:"blockHash" gorm:"type:CHAR(66)"`              //Block Hash
	BlockNumber types.Uint64   `json:"blockNumber" gorm:"index:idx_desc,sort:DESC"` //block number
	From        types.Address  `json:"from" gorm:"type:CHAR(42);index"`             //Send address
	To          *types.Address `json:"to" gorm:"type:CHAR(42);index"`               //Receive address
	Input       string         `json:"input"`                                       //Additional input data, contract call encoded data
	MethodId    *types.Data8   `json:"methodId,omitempty" gorm:"type:CHAR(18)"`     //Method ID, normal transaction is empty
	Value       types.BigInt   `json:"value" gorm:"type:VARCHAR(128)"`              //Amount, unit wei
	Nonce       types.Uint64   `json:"nonce"`                                       //Random number, the number of transactions initiated by the account
	Gas         types.Uint64   `json:"gas"`                                         // fuel
	GasPrice    types.Uint64   `json:"gasPrice"`                                    //Gas price
	Hash        types.Hash     `json:"hash" gorm:"type:CHAR(66);primaryKey"`        //Hash
	Receipt
}

// Receipt transaction receipt
type Receipt struct {
	Status            *types.Uint64  `json:"status,omitempty"`                                 //Status, 1: success; 0: failure
	CumulativeGasUsed types.Uint64   `json:"cumulativeGasUsed"`                                //Cumulative gas consumption
	ContractAddress   *types.Address `json:"contractAddress" gorm:"type:CHAR(42)"`             //The created contract address
	GasUsed           types.Uint64   `json:"gasUsed"`                                          //Gas consumption
	TxIndex           types.Uint64   `json:"transactionIndex" gorm:"index:idx_desc,sort:DESC"` //The serial number in the block
}

// Log transaction log
type Log struct {
	Address types.Address  `json:"address" gorm:"type:CHAR(42)"`                          //The contract address
	Topics  types.StrArray `json:"topics" gorm:"type:VARCHAR(277)"`                       //topic
	Data    string         `json:"data"`                                                  //data
	Removed bool           `json:"removed"`                                               //whether to remove
	TxHash  types.Hash     `json:"transactionHash" gorm:"type:CHAR(66);primaryKey;index"` //The transaction hash
	Index   types.Uint64   `json:"logIndex" gorm:"primaryKey"`                            //The serial number in the transaction
}

// InternalTx internal transaction
type InternalTx struct {
	ParentTxHash types.Hash     `json:"parentTxHash" gorm:"type:CHAR(66);index"` //The transaction
	Depth        types.Uint64   `json:"depth"`                                   //call depth
	Op           string         `json:"op" gorm:"type:VARCHAR(16)"`              //Operation
	From         *types.Address `json:"from" gorm:"type:CHAR(42);index"`         //Originating address
	To           *types.Address `json:"to" gorm:"type:CHAR(42);index"`           //Receive address
	Value        types.BigInt   `json:"value" gorm:"type:VARCHAR(128)"`          //Amount, unit wei
	GasLimit     types.Uint64   `json:"gasLimit"`                                //Gas limit
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
	Address       *string `json:"address" gorm:"type:CHAR(44);primary_key"`  //NFT address, grows automatically from 0x1
	RoyaltyRatio  uint32  `json:"royalty_ratio"`                             //Royalty rate, in ten thousandths
	MetaUrl       string  `json:"meta_url"`                                  //Real meta information URL
	RawMetaUrl    string  `json:"raw_meta_url"`                              //Original meta information URL on the chain
	ExchangerAddr string  `json:"exchanger_addr" gorm:"type:CHAR(44);index"` //The address of the exchange, if there is none, it can be traded on any exchange
	LastPrice     *string `json:"last_price"`                                //The last transaction price (null if the transaction is not completed), the unit is wei
	Creator       string  `json:"creator" gorm:"type:CHAR(44)"`              //Creator address
	Timestamp     uint64  `json:"timestamp"`                                 //Create timestamp
	BlockNumber   uint64  `json:"block_number"`                              //The height of the created block
	TxHash        string  `json:"tx_hash" gorm:"type:CHAR(66)"`              //The transaction hash created
	Owner         string  `json:"owner" gorm:"type:CHAR(44);index"`          //owner
	Name          string  `json:"name"`                                      //name
	Desc          string  `json:"desc"`                                      //description
	Attributes    string  `json:"attributes"`                                //Attributes
	Category      string  `json:"category"`                                  //category
	SourceUrl     string  `json:"source_url"`                                //Resource links, file links such as pictures or videos
	CollectionId  *string `json:"collection_id" gorm:"index"`                //The id of the collection, the name of the collection + the creator of the collection + the hash of the exchange where the collection is located
}

// Epoch SNFT Phase 1
// One SNFT->16 Collections->16 FNFTs->256 SNFTs
type Epoch struct {
	ID           string `json:"id" gorm:"type:VARCHAR(38);primary_key"` //period ID
	Creator      string `json:"creator" gorm:"type:CHAR(44)"`           //Creator address, also the address of royalty income
	RoyaltyRatio uint32 `json:"royaltyRatio"`                           //The royalty rate of the same period of SNFT, the unit is one ten thousandth
	Dir          string `json:"dir"`                                    //meta information directory URL
	Exchanger    string `json:"exchanger"`                              //Exchange address
	VoteWeight   string `json:"voteWeight"`                             //Weight
	Number       uint64 `json:"number"`                                 //Starting block height
	Timestamp    uint64 `json:"timestamp"`                              //Starting timestamp
}

// FNFT full SNFT
type FNFT struct {
	ID         string `json:"id" gorm:"type:VARCHAR(40);primary_key"` //FNFT ID
	MetaUrl    string `json:"meta_url"`                               //FNFT meta information URL
	Name       string `json:"name"`                                   //name
	Desc       string `json:"desc"`                                   //description
	Attributes string `json:"attributes"`                             //Attributes
	Category   string `json:"category"`                               //category
	SourceUrl  string `json:"source_url"`                             //Resource links, file links such as pictures or videos
}

// SNFT of SNFT fragments
type SNFT struct {
	Address      string  `json:"address" gorm:"type:CHAR(44);primary_key"` //SNFT address
	LastPrice    *string `json:"last_price"`                               //The last transaction price, the unit is wei, null if the transaction has not been completed
	Awardee      *string `json:"awardee"`                                  //The address of the miner that was rewarded last, null if it has not been rewarded
	RewardAt     *uint64 `json:"reward_at"`                                //The timestamp of the last rewarded, null if not rewarded
	RewardNumber *uint64 `json:"reward_number"`                            //The height of the last rewarded block, null if not rewarded
	Owner        *string `json:"owner" gorm:"type:CHAR(44);index"`         //owner, unallocated and reclaimed are null
}

// NFTTx NFT transaction attribute information
type NFTTx struct {
	//Transaction type, 1: transfer, 2: bid transaction, 3: fixed price purchase, 4: lazy price purchase, 5: lazy price purchase, 6: bid transaction, 7: lazy bid transaction, 8: matching transaction
	TxType        int32   `json:"tx_type"`
	NFTAddr       *string `json:"nft_addr" gorm:"type:CHAR(42);index"`      //The NFT address of the transaction
	ExchangerAddr string  `json:"exchanger_addr" gorm:"type:CHAR(42)"`      //Exchange address
	From          string  `json:"from" gorm:"type:CHAR(42);index"`          //Seller
	To            string  `json:"to" gorm:"type:CHAR(42);index"`            //buyer
	Price         *string `json:"price"`                                    //price, the unit is wei
	Timestamp     uint64  `json:"timestamp"`                                //Transaction timestamp
	TxHash        string  `json:"tx_hash" gorm:"type:CHAR(66);primary_key"` //transaction hash
	BlockNumber   uint64  `json:"block_number"`                             //block number
	Fee           *string `json:"fee"`                                      //Transaction fee, in wei (only if there is an exchange and price)
}

// Collection information, user collection: name + creator + hash of the exchange to which they belong, SNFT collection: SNFT address removes the last 3 digits
type Collection struct {
	Id          string  `json:"id" gorm:"type:CHAR(66);primary_key"`  //ID
	MetaUrl     string  `json:"meta_url"`                             //collection meta information URL
	Name        string  `json:"name"`                                 //name
	Creator     string  `json:"creator" gorm:"index"`                 //Creator
	Category    string  `json:"category"`                             //category
	Desc        string  `json:"desc"`                                 //description
	ImgUrl      string  `json:"img_url"`                              //image link
	BlockNumber uint64  `json:"block_number" gorm:"index"`            //Create block height, equal to the first NFT in the collection
	Exchanger   *string `json:"exchanger" gorm:"type:CHAR(42);index"` //belongs to the exchange
}

// Exchanger exchange attribute information
type Exchanger struct {
	Address      string `json:"address" gorm:"type:CHAR(42);primary_key"` //Exchange address
	Name         string `json:"name" gorm:"type:VARCHAR(256)"`            //Exchange name
	URL          string `json:"url"`                                      //Exchange URL
	FeeRatio     uint32 `json:"fee_ratio"`                                //fee rate, unit 1/10,000
	Creator      string `json:"creator" gorm:"type:CHAR(42)"`             //Creator address
	Timestamp    uint64 `json:"timestamp" gorm:"index"`                   //Open time
	BlockNumber  uint64 `json:"block_number" gorm:"index"`                //The block number when created
	TxHash       string `json:"tx_hash" gorm:"type:CHAR(66)"`             //The transaction created
	BalanceCount string `json:"balance_count" gorm:"type:VARCHAR(128)"`   //Total transaction amount, unit wei
	NFTCount     uint64 `json:"nft_count"`                                //Total NFT count
}

// Reward miner reward, the reward method is SNFT and Amount, and Amount is tentatively set to 0.1ERB
type Reward struct {
	Address     string  `json:"address" gorm:"type:CHAR(42)"` //reward address
	Proxy       string  `json:"proxy" gorm:"type:CHAR(42)"`   //proxy address
	Identity    uint8   `json:"identity"`                     //Identity, 1: block producer, 2: verifier, 3, exchange
	BlockNumber uint64  `json:"block_number"`                 //The block number when rewarding
	SNFT        *string `json:"snft" gorm:"type:CHAR(42)"`    //SNFT address
	Amount      *string `json:"amount"`                       //Amount of reward
}

// Pledge account pledge amount
type Pledge struct {
	Address string `json:"address" gorm:"type:CHAR(44);primary_key"` //staking account
	Amount  string `json:"amount" gorm:"type:CHAR(64)"`              //Pledge amount
	Count   uint64 `json:"count"`                                    //The number of pledges, both PledgeAdd and PledgeSub are added once
}

// ExchangerPledge exchange pledge
type ExchangerPledge Pledge

// ConsensusPledge consensus pledge
type ConsensusPledge Pledge

// Subscription information
type Subscription struct {
	Email string `json:"email" gorm:"type:VARCHAR(64);primary_key"` //Email
}
