package ethhelper

import (
	"encoding/hex"
	"errors"
	"fmt"
	"server/ethhelper/common"

	"math/big"
	"strings"
)

func IsApprovedNFT1155(owner, contract string) (bool, error) {
	var tmp big.Int
	payload := make([]byte, 68)
	tmp.SetString(common.IsApproveForAllHash, 0)
	copy(payload[:4], tmp.Bytes())
	tmp.SetString(owner, 0)
	copy(payload[36-len(tmp.Bytes()):], tmp.Bytes())
	tmp.SetString(common.TradeCore, 0)
	copy(payload[68-len(tmp.Bytes()):], tmp.Bytes())
	params := common.CallParamTemp{To: contract, Data: "0x" + hex.EncodeToString(payload)}
	var ret string
	if err := common.Client().Call(&ret, "eth_call", params, "latest"); err != nil {
		fmt.Println("IsApprovedNFT1155  Call failed:" + err.Error())
		return false, errors.New("IsApprovedNFT1155  Call failed:" + err.Error())
	} else {
		tmp.SetString(ret, 0)
		return tmp.Uint64() == 1, nil
	}
}

func IsApprovedNFT721(owner, contract, tokenId string) (bool, error) {
	var tmp big.Int
	payload := make([]byte, 36)
	tmp.SetString(common.GetApproved721Hash, 0)
	copy(payload[:4], tmp.Bytes())
	tmp.SetString(tokenId, 0)
	copy(payload[36-len(tmp.Bytes()):], tmp.Bytes())
	params := common.CallParamTemp{To: contract, Data: "0x" + hex.EncodeToString(payload)}
	var ret string
	if err := common.Client().Call(&ret, "eth_call", params, "latest"); err != nil {
		fmt.Println("IsApprovedNFT721  Call failed:" + err.Error())
		return false, errors.New("IsApprovedNFT721  Call failed:" + err.Error())
	} else {
		tmp.SetString(ret, 0)
		retHex := "0x" + strings.ToLower(hex.EncodeToString(tmp.Bytes()))
		if retHex != strings.ToLower(common.TradeCore) {
			return IsApprovedNFT1155(owner, contract)
		} else {
			return true, nil
		}
	}
}

func IsOwnerOfNFT721(owner, contract, tokenId string) (bool, error) {
	var tmp big.Int
	payload := make([]byte, 36)

	tmp.SetString(common.OwnerOf721Hash, 0)
	copy(payload[:4], tmp.Bytes())
	tmp.SetString(tokenId, 0)
	copy(payload[36-len(tmp.Bytes()):], tmp.Bytes())
	params := common.CallParamTemp{To: contract, Data: "0x" + hex.EncodeToString(payload)}
	var ret string
	if err := common.Client().Call(&ret, "eth_call", params, "latest"); err != nil {
		fmt.Println("IsOwnerOfNFT721  Call failed:" + err.Error())
		return false, errors.New("IsOwnerOfNFT721  Call failed:" + err.Error())
	} else {
		tmp.SetString(ret, 0)
		retHex := "0x" + strings.ToLower(hex.EncodeToString(tmp.Bytes()))
		return retHex == strings.ToLower(owner), nil
	}
}

func IsOwnerOfNFT1155(owner, contract, tokenId string) (bool, error) {
	var tmp big.Int
	payload := make([]byte, 64)
	//00fdd58e
	tmp.SetString(owner, 0)
	copy(payload[32-len(tmp.Bytes()):], tmp.Bytes())
	tmp.SetString(tokenId, 0)
	copy(payload[64-len(tmp.Bytes()):], tmp.Bytes())
	params := common.CallParamTemp{To: contract, Data: common.BalanceOf1155Hash + hex.EncodeToString(payload)}
	var ret string
	if err := common.Client().Call(&ret, "eth_call", params, "latest"); err != nil {
		fmt.Println("IsOwnerOfNFT721  Call failed:" + err.Error())
		return false, errors.New("IsOwnerOfNFT721  Call failed:" + err.Error())
	} else {
		tmp.SetString(ret, 0)
		return tmp.Uint64() > 0, nil
	}
}

func IsErc721(contract string) (int, error) {
	var tmp big.Int
	re := contractCall(contract, erc721Input)
	tmp.SetString(re, 0)
	if re == "" {
		return 0, errors.New("IsErc721: call failed")
	}
	if tmp.Uint64() != 1 {
		re = contractCall(contract, erc1155Input)
		if re == "" {
			return 0, errors.New("IsErc721: call failed")
		}
		tmp.SetString(re, 0)
		if tmp.Uint64() == 1 {
			return 2, nil
		}
		return 0, nil
	} else {
		return 1, nil
	}
}
