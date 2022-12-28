package types

import (
	"fmt"
	"math/big"
	"strconv"
)

type Long int64

// ImplementsGraphQLType returns true if Long implements the provided GraphQL type.
func (b *Long) ImplementsGraphQLType(name string) bool { return name == "Long" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (b *Long) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		// must be a hexadecimal string with a 0x prefix
		return b.UnmarshalText([]byte(input))
	case int32:
		*b = Long(input)
	case int64:
		*b = Long(input)
	default:
		err = fmt.Errorf("unexpected type %T for Long", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Long) UnmarshalJSON(input []byte) error {
	if len(input) > 2 && input[0] == '"' {
		input = input[1 : len(input)-1]
	}
	return b.UnmarshalText(input)
}

// UnmarshalText implements encoding.TextUnmarshaler
func (b *Long) UnmarshalText(input []byte) error {
	value, err := strconv.ParseInt(string(input), 0, 64)
	*b = Long(value)
	return err
}

func (b *Long) Hex() string {
	return "0x" + strconv.FormatUint(uint64(*b), 16)
}

type Bytes string

func (b *Bytes) ImplementsGraphQLType(name string) bool { return name == "Bytes" }

func (b *Bytes) UnmarshalGraphQL(input interface{}) error {
	if text, ok := input.(string); ok {
		*b = Bytes(text)
		return nil
	} else {
		return fmt.Errorf("unexpected type %T for Bytes", input)
	}
}

type Bytes8 string

// ImplementsGraphQLType returns true if Bytes8 implements the provided GraphQL type.
func (a *Bytes8) ImplementsGraphQLType(name string) bool { return name == "Bytes8" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (a *Bytes8) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return a.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Bytes8", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Bytes8) UnmarshalJSON(input []byte) error {
	return a.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (a *Bytes8) UnmarshalText(input []byte) error {
	*a = Bytes8(input)
	return nil
}

type Address string

// ImplementsGraphQLType returns true if Address implements the provided GraphQL type.
func (a *Address) ImplementsGraphQLType(name string) bool { return name == "Address" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (a *Address) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return a.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Address", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Address) UnmarshalJSON(input []byte) error {
	return a.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (a *Address) UnmarshalText(input []byte) error {
	*a = Address(input)
	return nil
}

type Hash string

// ImplementsGraphQLType returns true if Hash implements the provided GraphQL type.
func (b *Hash) ImplementsGraphQLType(name string) bool { return name == "Hash" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (b *Hash) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return b.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Address", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Hash) UnmarshalJSON(input []byte) error {
	return b.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (b *Hash) UnmarshalText(input []byte) error {
	*b = Hash(input)
	return nil
}

// BigInt big number represented by decimal string
type BigInt string

// ImplementsGraphQLType returns true if BigInt implements the provided GraphQL type.
func (b *BigInt) ImplementsGraphQLType(name string) bool { return name == "BigInt" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (b *BigInt) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return b.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for BigInt", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *BigInt) UnmarshalJSON(input []byte) error {
	return b.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (b *BigInt) UnmarshalText(input []byte) error {
	t := new(big.Int)
	err := t.UnmarshalText(input)
	if err != nil {
		return err
	}
	*b = BigInt(t.String())
	return nil
}

type ContractType int

const (
	ERC20 ContractType = iota + 1
	ERC165
	ERC721
	ERC1155
)

// ImplementsGraphQLType returns true if Long implements the provided GraphQL type.
func (e *ContractType) ImplementsGraphQLType(name string) bool { return name == "ContractType" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (e *ContractType) UnmarshalGraphQL(input interface{}) error {
	if text, ok := input.(string); ok {
		switch text {
		case "ERC20":
			*e = ERC20
		case "ERC165":
			*e = ERC165
		case "ERC721":
			*e = ERC721
		case "ERC1155":
			*e = ERC1155
		}
		return nil
	} else {
		return fmt.Errorf("unexpected type %T for ContractType", input)
	}
}

// MarshalJSON implements json.Marshaler.
func (e *ContractType) MarshalJSON() ([]byte, error) {
	switch *e {
	case ERC20:
		return []byte("\"ERC20\""), nil
	case ERC165:
		return []byte("\"ERC165\""), nil
	case ERC721:
		return []byte("\"ERC721\""), nil
	case ERC1155:
		return []byte("\"ERC1155\""), nil
	default:
		return []byte("\"NONE\"")[:], nil
	}
}
