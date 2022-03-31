package block

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"server/log"
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	sigTime := fmt.Sprintf("%v", time.Now().Unix())
	_, d := GenMintBNBSign("0x5B38Da6a701c568545dCfcB03FcB875f56beddC4", "100000000000000000", "1", sigTime)
	fmt.Println(d)
	_, d = GenMintNFTSign("0x5B38Da6a701c568545dCfcB03FcB875f56beddC4", "1", "13", "12", "123456", sigTime)
	fmt.Println(d)

	_, d = genConsumeFSK("0xB04cC2abA7e5626a1216A1C9C60255e06cAC0e04", "2000", "1646882647")
	fmt.Println(d, sigTime)
}

const VerifyPrivateKey = "864b5cd0869d4a8c0e432a2d8d05d2f95fbe6572104228d16aeaa85b2a3edc8c"

//ConsumeFSK 签名sig
func genConsumeFSK(address, tokenBalance, sigTime string) (error, string) {
	toBytes := func(v string) []byte {
		var bigTemp big.Int
		bigTemp.SetString(v, 0)
		return bigTemp.Bytes()
	}
	data := crypto.Keccak256Hash(
		common.HexToAddress(address).Bytes(),
		common.LeftPadBytes(toBytes(tokenBalance), 32),
		common.LeftPadBytes(toBytes(sigTime), 32),
	)

	sign, err := Sign(data)
	if err != nil {
		log.Info("GenMintBNBSign() sign err=", err)
		return err, ""
	}
	return nil, sign
}

//MintBnb 签名sig
//address(授权调用MintBnb的地址)  bnbAmount(授权的bnb数量，需要处理精度,按wei的单位输入) types(类型)   sigTime(授权签名时的时间戳，精确到秒)
func GenMintBNBSign(address, bnbAmount, types, sigTime string) (error, string) {
	toBytes := func(v string) []byte {
		var bigTemp big.Int
		bigTemp.SetString(v, 0)
		return bigTemp.Bytes()
	}
	data := crypto.Keccak256Hash(
		common.HexToAddress(address).Bytes(),
		common.LeftPadBytes(toBytes(bnbAmount), 32),
		toBytes(types),
		common.LeftPadBytes(toBytes(sigTime), 32),
	)

	sign, err := Sign(data)
	if err != nil {
		log.Info("GenMintBNBSign() sign err=", err)
		return err, ""
	}
	return nil, sign
}

//MintNFT 签名sig
//address(授权调用MintNFT的地址) isgenerate(true则传1,false则传0), rare, fromType, production (这些字段与合约调用参数一致) sigTime(授权签名时的时间戳，精确到秒)
func GenMintNFTSign(address, isgenerate, rare, fromType, production, sigTime string) (error, string) {
	toBytes := func(v string) []byte {
		var bigTemp big.Int
		bigTemp.SetString(v, 0)
		return bigTemp.Bytes()
	}
	data := crypto.Keccak256Hash(
		common.HexToAddress(address).Bytes(),
		toBytes(isgenerate),
		toBytes(rare),
		toBytes(fromType),
		common.LeftPadBytes(toBytes(production), 32),
		common.LeftPadBytes(toBytes(sigTime), 32),
	)

	sign, err := Sign(data)
	if err != nil {
		log.Info("GenMintBNBSign() sign err=", err)
		return err, ""
	}
	return nil, sign
}

func Sign(data common.Hash) (string, error) {
	key, err := crypto.HexToECDSA(VerifyPrivateKey)
	if err != nil {
		log.Info("Sign() key err=", err)
		return "", err
	}
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n32%s", data.Bytes())
	sig, err := crypto.Sign(crypto.Keccak256([]byte(msg)), key)
	if err != nil {
		log.Info("Sign signature error: ", err)
		return "", err
	}
	sig[64] += 27
	return hexutil.Encode(sig), err
}
