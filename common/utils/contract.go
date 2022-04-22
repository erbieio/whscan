package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrNoCode   = errors.New("no contract code at given address")
	DefaultOpts = &bind.CallOpts{}
)

func Call(abi abi.ABI, address common.Address, caller bind.ContractCaller, opts *bind.CallOpts, results *[]interface{}, method string, params ...interface{}) error {
	// Don't crash on a lazy user
	if opts == nil {
		opts = new(bind.CallOpts)
	}
	if results == nil {
		results = new([]interface{})
	}
	// Pack the input, call and unpack the results
	input, err := abi.Pack(method, params...)
	if err != nil {
		return err
	}
	var (
		msg    = ethereum.CallMsg{From: opts.From, To: &address, Data: input}
		ctx    = ensureContext(opts.Context)
		code   []byte
		output []byte
	)

	output, err = caller.CallContract(ctx, msg, opts.BlockNumber)
	if err != nil {
		return err
	}
	if len(output) == 0 {
		// Make sure we have a contract to operate on, and bail out otherwise.
		if code, err = caller.CodeAt(ctx, address, opts.BlockNumber); err != nil {
			return err
		} else if len(code) == 0 {
			return ErrNoCode
		}
	}

	if len(*results) == 0 {
		res, err := abi.Unpack(method, output)
		*results = res
		return err
	}
	res := *results
	return abi.UnpackIntoInterface(res[0], method, output)
}

type Log struct {
	Topics []common.Hash
	Data   []byte
}

func UnpackLog(a abi.ABI, out interface{}, event string, log Log) error {
	if log.Topics[0] != a.Events[event].ID {
		return fmt.Errorf("event signature mismatch")
	}
	if len(log.Data) > 0 {
		if err := a.UnpackIntoInterface(out, event, log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range a.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}

// FilterContractErr 过滤掉除网络连接外的错误
func FilterContractErr(err error) error {
	if err != nil {
		if strings.Index(err.Error(), "connection") > 0 {
			return err
		}
		if strings.Index(err.Error(), "unexpected EOF") > 0 {
			return err
		}
	}
	return nil
}

// ensureContext is a helper method to ensure a context is not nil, even if the
// user specified it as such.
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
