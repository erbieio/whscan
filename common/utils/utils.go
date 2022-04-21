package utils

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// Sign 签名数据
func Sign(data []byte, key *ecdsa.PrivateKey) (string, error) {
	sig, err := crypto.Sign(accounts.TextHash(data), key)
	if err != nil {
		return "", err
	}
	sig[64] += 27
	return hexutil.Encode(sig), err
}

// VerifyDateSig 验证对当前和昨天UTC日期（如：20220404）签名
func VerifyDateSig(hexSig, addr string) bool {
	// todo 每次请求都要计算nowHash和oldHash，可优化为后台定时计算
	now := time.Now()
	nowHash := accounts.TextHash([]byte(now.UTC().Format("20060102")))
	oldHash := accounts.TextHash([]byte(now.Add(-24 * time.Hour).UTC().Format("20060102")))
	sig, err := hexutil.Decode(hexSig)
	if err != nil {
		return false
	}
	if len(sig) != 65 {
		return false
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	address := common.HexToAddress(addr)
	rpk, err := crypto.SigToPub(nowHash, sig)
	if err != nil {
		return false
	}
	if crypto.PubkeyToAddress(*rpk) == address {
		return true
	}
	rpk, err = crypto.SigToPub(oldHash, sig)
	if err != nil {
		return false
	}
	return crypto.PubkeyToAddress(*rpk) == address
}

func ToBytes(v string) []byte {
	var bigTemp big.Int
	bigTemp.SetString(v, 0)
	return bigTemp.Bytes()
}

// ParsePage 解析分页参数，默认值是第一页10条记录
func ParsePage(pagePtr, sizePtr *int) (int, int, error) {
	page, size := 1, 10
	if pagePtr != nil {
		if *pagePtr <= 0 {
			return 0, 0, errors.New("分页页数小于1")
		}
		page = *pagePtr
	}
	if sizePtr != nil {
		if *sizePtr <= 0 {
			return 0, 0, errors.New("分页大小小于1")
		}
		size = *sizePtr
	}
	return page, size, nil
}
