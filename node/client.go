package node

import (
	"context"

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

func (c *Client) BlockNumber(ctx context.Context) (result types.Long, err error) {
	err = c.CallContext(ctx, &result, "eth_blockNumber")
	return
}

func (c *Client) CallContract(ctx context.Context, to, data, number any) (result types.Bytes, err error) {
	err = c.CallContext(ctx, &result, "eth_call", map[string]any{"to": to, "data": data}, number)
	return
}
