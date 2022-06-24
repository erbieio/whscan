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

// Data 用带前缀16进制字符串表示的字节数组
type Data string

// ImplementsGraphQLType returns true if Data implements the provided GraphQL type.
func (d Data) ImplementsGraphQLType(name string) bool { return name == "Data" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (d *Data) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return d.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Data", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Data) UnmarshalJSON(input []byte) error {
	return d.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Data) UnmarshalText(input []byte) error {
	*d = Data(input)
	return nil
}

type Data8 string

// ImplementsGraphQLType returns true if Data8 implements the provided GraphQL type.
func (a Data8) ImplementsGraphQLType(name string) bool { return name == "Data8" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (a *Data8) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return a.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Data8", input)
	}
	return err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Data8) UnmarshalJSON(input []byte) error {
	return a.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (a *Data8) UnmarshalText(input []byte) error {
	*a = Data8(input)
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

// BigInt 包装big.Int，增加数据库和json相关函数
type BigInt struct{ big.Int }

func (b BigInt) ImplementsGraphQLType(name string) bool { return name == "BigInt" }

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

func (b *BigInt) UnmarshalJSON(input []byte) error {
	return b.UnmarshalText(input[1 : len(input)-1])
}

func (b *BigInt) Scan(src interface{}) error {
	switch srcT := src.(type) {
	case string:
		return b.UnmarshalText([]byte(srcT))
	case []byte:
		return b.UnmarshalText(srcT)
	default:
		return fmt.Errorf("can't scan %T into BigInt", src)
	}
}

func (b *BigInt) Value() (driver.Value, error) {
	return b.Text(10), nil
}

func (b *BigInt) Hex() string {
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
