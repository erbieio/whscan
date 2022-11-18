package utils

import (
	"fmt"
	"math/big"
	"testing"

	"server/common/model"
)

func TestBigToAddress(t *testing.T) {
	a := big.NewInt(1)
	t.Log(BigToAddress(a))
	t.Log(BigToAddress(nil))
	a.SetString("bbbbbbbbbbbbbbbbbbbbbbbbaaaaaaaaaaaaaaaaaaaaaaaa", 16)
	t.Log(BigToAddress(a))
	t.Log(a.SetString("0aabbb", 16))
	t.Log(fmt.Sprintf("%s%x", "0x1234", 0))
	t.Log(fmt.Sprintf("%s%x", "0x1234", 13))
	t.Log(IP2Location("4.4.4.4"))
}

func TestKeccak256Hash(t *testing.T) {
	t.Log(Keccak256Hash([]byte("Transfer(address,address,uint256)")))
	t.Log(Keccak256Hash([]byte("Transfer(address,address,uint256)")))
	t.Log(Keccak256Hash([]byte("TransferSingle(address,address,address,uint256,uint256)")))
	t.Log(Keccak256Hash([]byte("TransferBatch(address,address,address,uint256[],uint256[])")))

	t.Log(Keccak256Hash([]byte("name()"))[:10])
	t.Log(Keccak256Hash([]byte("symbol()"))[:10])
	t.Log(Keccak256Hash([]byte("decimals()"))[:10])
	t.Log(Keccak256Hash([]byte("totalSupply()"))[:10])

	t.Log(Keccak256Hash([]byte("allowance(address,address)"))[:10])
	t.Log(Keccak256Hash([]byte("balanceOf(address)"))[:10])

	t.Log(Keccak256Hash([]byte("transfer(address,uint256)"))[:10])
	t.Log(Keccak256Hash([]byte("transferFrom(address,address,uint256)"))[:10])
	t.Log(Keccak256Hash([]byte("approve(address,uint256)"))[:10])
}

func TestPubkeyToAddress(t *testing.T) {
	key, _ := HexToECDSA("7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd")
	t.Log(PubkeyToAddress(key.PubKey()))
	// 0x394586580ff4170c8a0244837202cbabe9070f66
	msg := "hello"
	hash := Keccak256([]byte(msg))
	sig, err := Sign(hash[:], key)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%x", sig)
	pub, err := SigToPub(hash, sig)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(PubkeyToAddress(pub))
}

func TestUnpack1155TransferLog(t *testing.T) {
	log := model.EventLog{
		Topics: []string{"0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb", "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb", "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb", "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb"},
		Data:   "0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000008",
	}
	transferLog := UnpackTransferLog(&log)
	t.Logf("%+v%+v", transferLog[0], transferLog[1])
}
