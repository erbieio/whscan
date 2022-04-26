package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
)

type Uint64 uint64

// ImplementsGraphQLType returns true if Long implements the provided GraphQL type.
func (b Uint64) ImplementsGraphQLType(name string) bool { return name == "Long" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (b *Uint64) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		// 必须是带0x前缀的16进制字符串
		return b.UnmarshalText([]byte(input))
	case int32:
		*b = Uint64(input)
	case int64:
		*b = Uint64(input)
	default:
		err = fmt.Errorf("unexpected type %T for Uint64", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Uint64) UnmarshalJSON(input []byte) error {
	return b.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (b *Uint64) UnmarshalText(input []byte) error {
	value, err := strconv.ParseUint(string(input), 0, 64)
	*b = Uint64(value)
	return err
}

func (b Uint64) String() string {
	return b.Hex()
}

func (b Uint64) Hex() string {
	return "0x" + strconv.FormatUint(uint64(b), 16)
}

type Bytes string

// ImplementsGraphQLType returns true if Bytes implements the provided GraphQL type.
func (a Bytes) ImplementsGraphQLType(name string) bool { return name == "Bytes" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (a *Bytes) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return a.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Bytes", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Bytes) UnmarshalJSON(input []byte) error {
	return a.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (a *Bytes) UnmarshalText(input []byte) error {
	*a = Bytes(input)
	return nil
}

type Bytes4 string

// ImplementsGraphQLType returns true if Bytes4 implements the provided GraphQL type.
func (a Bytes4) ImplementsGraphQLType(name string) bool { return name == "Bytes4" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (a *Bytes4) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return a.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Bytes4", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Bytes4) UnmarshalJSON(input []byte) error {
	return a.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (a *Bytes4) UnmarshalText(input []byte) error {
	*a = Bytes4(input)
	return nil
}

type Bytes8 string

// ImplementsGraphQLType returns true if Bytes8 implements the provided GraphQL type.
func (a Bytes8) ImplementsGraphQLType(name string) bool { return name == "Bytes8" }

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
func (a Address) ImplementsGraphQLType(name string) bool { return name == "Address" }

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
func (b Hash) ImplementsGraphQLType(name string) bool { return name == "Hash" }

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

type Uint256 string

var Big0 Uint256 = "0x0000000000000000000000000000000000000000000000000000000000000000"
var Big1 Uint256 = "0x0000000000000000000000000000000000000000000000000000000000000001"

// ImplementsGraphQLType returns true if Uint256 implements the provided GraphQL type.
func (a Uint256) ImplementsGraphQLType(name string) bool { return name == "Uint256" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (a *Uint256) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return a.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Uint256", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Uint256) UnmarshalJSON(input []byte) error {
	return a.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (a *Uint256) UnmarshalText(input []byte) error {
	*a = Uint256(input)
	return nil
}

// BigInt 由十进制字符串方式表示的大数
type BigInt string

// ImplementsGraphQLType returns true if BigInt implements the provided GraphQL type.
func (a BigInt) ImplementsGraphQLType(name string) bool { return name == "BigInt" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (a *BigInt) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return a.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for BigInt", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *BigInt) UnmarshalJSON(input []byte) error {
	return a.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (a *BigInt) UnmarshalText(input []byte) error {
	if input[0] == '0' && input[1] == 'x' || input[1] == 'X' {
		// 16进制转10进制
		b := new(big.Int)
		err := b.UnmarshalText(input)
		if err != nil {
			return err
		}
		*a = BigInt(b.String())
	} else {
		*a = BigInt(input)
	}
	return nil
}

func (a BigInt) Hex() string {
	b := new(big.Int)
	b.SetString(string(a), 10)
	return "0x" + b.Text(16)
}

type StrArray []string

// Scan implements Scanner for database/sql.
func (sa *StrArray) Scan(src interface{}) error {
	srcS, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into string", src)
	}
	return json.Unmarshal(srcS, &sa)
}

// Value implements valuer for database/sql.
func (sa StrArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "[]", nil
	}
	enc, err := json.Marshal(sa)
	return string(enc), err
}

type ERC int

const (
	NONE ERC = iota
	ERC20
	ERC165
	ERC721
	ERC1155
)

// ImplementsGraphQLType returns true if Long implements the provided GraphQL type.
func (e ERC) ImplementsGraphQLType(name string) bool { return name == "ERC" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (e *ERC) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	default:
		err = fmt.Errorf("unexpected type %T for ERC", input)
	}
	return err
}

// MarshalJSON implements json.Marshaler.
func (e ERC) MarshalJSON() ([]byte, error) {
	switch e {
	case NONE:
		return []byte("\"NONE\"")[:], nil
	case ERC20:
		return []byte("\"ERC20\""), nil
	case ERC165:
		return []byte("\"ERC165\""), nil
	case ERC721:
		return []byte("\"ERC721\""), nil
	case ERC1155:
		return []byte("\"ERC1155\""), nil
	default:
		return nil, fmt.Errorf("unexpected value %v for contract type", e)
	}
}
