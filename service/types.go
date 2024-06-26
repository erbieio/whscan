package service

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// ErrRes interface error message returned
type ErrRes struct {
	ErrStr string `json:"err_str"` //Error message
}

// BigIntAdd adds two large number strings and returns a decimal large number string
func BigIntAdd(a, b string) string {
	aa, ok := new(big.Int).SetString(a, 0)
	if !ok {
		panic(nil)
	}
	bb, ok := new(big.Int).SetString(b, 0)
	if !ok {
		panic(nil)
	}
	cc := aa.Add(aa, bb)
	if cc.Sign() == -1 {
		panic(fmt.Sprintf("error: the result is negative number, %s+%s=%s", a, b, cc.String()))
	}
	return cc.Text(10)
}

func TxFee(price string, ratio int64) *string {
	value, ok := new(big.Int).SetString(price, 0)
	if !ok {
		return nil
	}
	value = value.Mul(value, big.NewInt(ratio))
	if value.Sign() == -1 {
		return nil
	}
	fee := value.Div(value, big.NewInt(10000)).Text(10)
	return &fee
}

var minValidatorAmount, _ = new(big.Int).SetString("35000000000000000000000", 10)

func CheckValidatorAmount(amount string) bool {
	value, ok := new(big.Int).SetString(amount, 0)
	if !ok {
		return false
	}
	return value.Cmp(minValidatorAmount) != -1
}

var LevelValues = []*big.Float{
	big.NewFloat(30000000000000000),
	big.NewFloat(60000000000000000),
	big.NewFloat(180000000000000000),
	big.NewFloat(1000000000000000000),
}

func _value(hexEpoch string, level int, pieces int64) *big.Int {
	epoch, err := strconv.ParseInt(hexEpoch, 16, 64)
	if err != nil {
		panic(err)
	}
	value := new(big.Float).Mul(LevelValues[level], big.NewFloat(float64(pieces)))
	if year := epoch / 6160; year > 0 {
		value.Mul(value, big.NewFloat(math.Pow(0.88, float64(year))))
	}
	result, _ := value.Int(nil)
	return result
}

func snftValue(snft string, pieces int64) string {
	return _value(snft[3:39], 42-len(snft), pieces).Text(10)
}
