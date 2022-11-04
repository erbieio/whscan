package node

import (
	"context"
	"math/big"
	"strconv"

	"server/common/model"
	"server/common/types"
)

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

func (c *Client) ChainId() (result types.Uint64, err error) {
	err = c.Call(&result, "eth_chainId")
	return
}

func (c *Client) Genesis() (result *model.Header, err error) {
	err = c.Call(&result, "eth_getBlockByNumber", "0x0", false)
	return
}

func (c *Client) BlockNumber() (result types.Uint64, err error) {
	err = c.Call(&result, "eth_blockNumber")
	return
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

func (c *Client) IsDebug() bool {
	return c.Call(&struct{}{}, "debug_gcStats") == nil
}

func (c *Client) IsWormholes() bool {
	return c.Call(&struct{}{}, "eth_getAccountInfo", "0x0000000000000000000000000000000000000000", "0x0") == nil
}

type Epoch struct {
	Dir        string   `json:"dir"`
	Royalty    uint32   `json:"royalty"`
	Creator    string   `json:"creator"`
	Address    string   `json:"address"` //Exchange address
	VoteWeight *big.Int `json:"vote_weight"`
}

func (c *Client) GetEpoch(number string) (*Epoch, error) {
	var epoches = struct {
		InjectedOfficialNFTs []*Epoch `json:"InjectedOfficialNFTs"`
	}{}
	err := c.Call(&epoches, "eth_getInjectedNFTInfo", number)
	if err != nil {
		return nil, err
	}
	if len(epoches.InjectedOfficialNFTs) == 2 {
		return epoches.InjectedOfficialNFTs[1], nil
	} else {
		return epoches.InjectedOfficialNFTs[0], nil
	}
}
