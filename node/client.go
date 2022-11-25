package node

import (
	"context"
	"math/big"
	"strconv"

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

func (c *Client) ChainId() (result types.Long, err error) {
	err = c.Call(&result, "eth_chainId")
	return
}

func (c *Client) BlockNumber(ctx context.Context) (result types.Long, err error) {
	err = c.CallContext(ctx, &result, "eth_blockNumber")
	return
}

func (c *Client) CallContract(ctx context.Context, to, data, number any) (result types.Bytes, err error) {
	err = c.CallContext(ctx, &result, "eth_call", map[string]any{"to": to, "data": data}, number)
	return
}
