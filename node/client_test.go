package node

import (
	"testing"

	"server/common/utils"
)

func TestClient_GetERC(t *testing.T) {
	client, err := Dial("http://127.0.0.1:8545")
	if err != nil {
		t.Error(err)
		return
	}
	erc, err := utils.IsERC20(client, "0x0a057a87ce9c56d7e336b417c79cf30e8d27860b")
	if err != nil {
		t.Error(err)
	}
	t.Log(erc)
}
