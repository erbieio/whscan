package utils

import (
	"server/common/model"

	"math/big"
	"testing"
)

func TestBigToAddress(t *testing.T) {
	a := big.NewInt(1)
	t.Log(BigToAddress(a))
	t.Log(BigToAddress(nil))
	a.SetString("bbbbbbbbbbbbbbbbbbbbbbbbaaaaaaaaaaaaaaaaaaaaaaaa", 16)
	t.Log(BigToAddress(a))
}

func TestKeccak256Hash(t *testing.T) {
	t.Log(Keccak256Hash([]byte("Transfer(address,address,uint256)")))
	t.Log(Keccak256Hash([]byte("Transfer(address,address,uint256)")))
	t.Log(Keccak256Hash([]byte("TransferSingle(address,address,address,uint256,uint256)")))
	t.Log(Keccak256Hash([]byte("TransferBatch(address,address,address,uint256[],uint256[])")))
}

func TestPubkeyToAddress(t *testing.T) {
	key, _ := HexToECDSA("7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd")
	t.Log(PubkeyToAddress(key.PublicKey))
	// 0x394586580ff4170c8a0244837202cbabe9070f66
}

func TestUnpack1155TransferLog(t *testing.T) {
	log := model.Log{
		Topics: []string{"0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb", "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb", "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb", "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb"},
		Data:   "0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000008",
	}
	transferLog, err := Unpack1155TransferLog(&log)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v%+v", transferLog[0], transferLog[1])
	}
}
