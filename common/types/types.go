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
		// must be a hexadecimal string with a 0x prefix
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
	if len(input) > 2 && input[0] == '"' {
		input = input[1 : len(input)-1]
	}
	return b.UnmarshalText(input)
}

// UnmarshalText implements encoding.TextUnmarshaler
func (b *Uint64) UnmarshalText(input []byte) error {
	value, err := strconv.ParseUint(string(input), 0, 64)
	*b = Uint64(value)
	return err
}

func (b Uint64) Hex() string {
	return "0x" + strconv.FormatUint(uint64(b), 16)
}

// Data is a byte array represented by a prefixed hex string
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

// BigInt big number represented by decimal string
type BigInt string

// ImplementsGraphQLType returns true if BigInt implements the provided GraphQL type.
func (b BigInt) ImplementsGraphQLType(name string) bool { return name == "BigInt" }

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

func (b BigInt) Hex() string {
	t := new(big.Int)
	t.SetString(string(b), 10)
	return "0x" + t.Text(16)
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
