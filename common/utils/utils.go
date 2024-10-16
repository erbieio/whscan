package utils

import (
	"fmt"
	"math/big"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	. "server/common/model"
	. "server/common/types"
)

var (
	erc20TransferEventId         = Hash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	erc721TransferEventId        = Hash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	erc1155TransferSingleEventId = Hash("0x7b912cc6629daab379d004780e875cdb7625e8331d3a7c8fbe08a42156325546")
	erc1155TransferBatchEventId  = Hash("0x20114eb39ee5dfdb13684c7d9e951052ef22c89bff67131a9bf08879189b0f71")
	loc, _                       = time.LoadLocation("Local")
	DaySecond                    = 24 * time.Hour.Milliseconds() / 1000
)

func UnpackTransferLog(log *EventLog) []any {
	topicsLen := len(log.Topics)
	if topicsLen == 3 {
		if log.Topics[0] == erc20TransferEventId && len(log.Data) == 66 {
			// Parse ERC20 transition events
			return []any{&ERC20Transfer{
				TxHash:  log.TxHash,
				Address: log.Address,
				From:    Address("0x" + log.Topics[1][26:]),
				To:      Address("0x" + log.Topics[2][26:]),
				Value:   HexToBigInt(log.Data[2:66]),
			}}
		}
	} else if topicsLen == 4 {
		if log.Topics[0] == erc721TransferEventId && len(log.Data) == 2 {
			// Parse ERC721 transition events
			return []any{&ERC721Transfer{
				TxHash:  log.TxHash,
				Address: log.Address,
				From:    Address("0x" + log.Topics[1][26:]),
				To:      Address("0x" + log.Topics[2][26:]),
				TokenId: HexToBigInt(string(log.Topics[3][2:])),
			}}
		} else if log.Topics[0] == erc1155TransferSingleEventId && len(log.Data) == 130 {
			// Parse ERC1155 transition events
			operator, from, to := Address("0x"+log.Topics[1][26:]), Address("0x"+log.Topics[2][26:]), Address("0x"+log.Topics[3][26:])
			return []any{&ERC1155Transfer{
				TxHash:   log.TxHash,
				Address:  log.Address,
				Operator: operator,
				From:     from,
				To:       to,
				TokenId:  HexToBigInt(log.Data[2:66]),
				Value:    HexToBigInt(log.Data[66:130]),
			}}
		} else if log.Topics[0] == erc1155TransferBatchEventId {
			// Parse the batch transfer events of ERC1155
			operator, from, to := Address("0x"+log.Topics[1][26:]), Address("0x"+log.Topics[2][26:]), Address("0x"+log.Topics[3][26:])
			// Dynamic data type codec reference: https://docs.soliditylang.org/en/v0.8.13/abi-spec.html#argument-encoding
			// The word length is 256 bits, or 32 bytes
			wordLen := (len(log.Data) - 2) / 64
			if wordLen < 4 || wordLen%2 != 0 || log.Data[2:66] != "0000000000000000000000000000000000000000000000000000000000000040" {
				return nil
			}
			transferCount := (wordLen - 4) / 2
			transferLogs := make([]any, transferCount)
			for i := 0; i < transferCount; i++ {
				idOffset, valueOffset := 2+(i+3)*64, 2+(transferCount+i+4)*64
				transferLogs[i] = &ERC1155Transfer{
					TxHash:   log.TxHash,
					Address:  log.Address,
					Operator: operator,
					From:     from,
					To:       to,
					TokenId:  HexToBigInt(log.Data[idOffset : idOffset+64]),
					Value:    HexToBigInt(log.Data[valueOffset : valueOffset+64]),
				}
			}
			return transferLogs
		}
	}
	return nil
}

// HexToBigInt converts a hexadecimal string without 0x prefix to a big number BigInt (illegal input will return 0)
func HexToBigInt(hex string) BigInt {
	b := new(big.Int)
	b.SetString(hex, 16)
	return BigInt(b.Text(10))
}

// FormatAddress converts a hexadecimal string without a 0x prefix to an Address (greater than the truncated front)
func FormatAddress(hex string) *Address {
	if len(hex) < 42 {
		hex = "0x000000000000000000000000000000000000000" + hex[2:]
	}
	if len(hex) > 42 {
		hex = "0x" + hex[len(hex)-40:]
	}
	return (*Address)(&hex)
}

// ParseAddress converts a hexadecimal string prefixed with 0x to an address
func ParseAddress(hex []byte) (Address, error) {
	if len(hex) != 42 {
		return "", fmt.Errorf("length is not 42")
	}
	if hex[0] != '0' || (hex[1] != 'x' && hex[1] != 'X') {
		return "", fmt.Errorf("prefix is not 0x")
	}
	hex[1] = 'x'
	for i := 2; i < 42; i++ {
		c := hex[i]
		if ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') {
			continue
		}
		if 'A' <= c && c <= 'F' {
			hex[i] = c + 32
			continue
		}
		return "", fmt.Errorf("illegal character: %v", c)
	}
	return Address(hex), nil
}

// BigToAddress converts large numbers into addresses (too large numbers will truncate the previous ones)
func BigToAddress(big *big.Int) Address {
	addr := "0000000000000000000000000000000000000000"
	if big != nil {
		addr += big.Text(16)
	}
	return Address("0x" + addr[len(addr)-40:])
}

func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := HomeDir(); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func NewWatcher(files []string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		err = watcher.Add(file)
		if err != nil {
			return nil, err
		}
	}
	return watcher, nil
}

// ParsePagination Parsing pagination parameters, maximum 100 records, default return 10 records on page 1
func ParsePagination(pageStr, sizeStr string) (int, int) {
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 || size > 100 {
		size = 10
	}

	return page, size
}

// LastTimeRange unix time range for the specified number of days based on the current time
func LastTimeRange(day int64) (start, stop int64) {
	now := time.Now().Local()
	stopTime, _ := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02")+" 00:00:00", loc)
	stop = stopTime.Unix()
	start = stop - DaySecond*day
	return
}

// LastTimeRange2 unix time range for the specified number of days based on the current time
func LastTimeRange2(day int64) (start, stop int64) {
	stop = time.Now().Unix()
	start = stop - DaySecond*day
	return
}
