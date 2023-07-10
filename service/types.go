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

func snftMergeValue(snft string, pieces int64) string {
	value := _value(snft[3:39], 42-len(snft), pieces)
	_value := _value(snft[3:39], 41-len(snft), pieces)
	return value.Sub(value, _value).Text(10)
}

func apr(rewardSNFT int64, totalPledge string) float64 {
	reward := big.NewFloat(12 * 60 * 24 * 365 * 0.77)
	pledge, _ := new(big.Float).SetString(totalPledge[:len(totalPledge)-18])
	if year := rewardSNFT / 4096 / 6160; year > 0 {
		reward.Mul(reward, big.NewFloat(math.Pow(0.88, float64(year))))
	}
	result, _ := reward.Quo(reward, pledge).Float64()
	return result
}
