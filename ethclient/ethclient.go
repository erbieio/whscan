package ethclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/INFURA/go-ethlibs/jsonrpc"
	"github.com/INFURA/go-ethlibs/node"
	"server/common/types"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client struct {
	node.Client
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	client, err := node.NewClient(context.Background(), rawurl)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func (c *Client) Call(result interface{}, method string, args ...interface{}) error {
	ctx := context.Background()
	return c.CallContext(ctx, result, method, args...)
}

func (c *Client) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: method,
		Params: jsonrpc.MustParams(args...),
	}

	response, err := c.Request(ctx, &request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return errors.New(string(*response.Error))
	}
	return json.Unmarshal(response.Result, &result)
}

type BatchElem struct {
	Method string
	Args   []interface{}
	Result interface{}
	Error  error
}

func (c *Client) BatchCall(b []BatchElem) error {
	ctx := context.Background()
	return c.BatchCallContext(ctx, b)
}

func (c *Client) BatchCallContext(ctx context.Context, b []BatchElem) error {
	for _, elem := range b {
		elem.Error = c.CallContext(ctx, elem.Result, elem.Method, elem.Args...)
	}
	return nil
}

func (c *Client) CallContract(ctx context.Context, msg map[string]interface{}, blockNumber *types.BigInt) (types.Bytes, error) {
	var hex types.Bytes
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
