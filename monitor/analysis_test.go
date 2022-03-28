package monitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
	"math/big"
	"testing"
)

func TestBalancseOf(t *testing.T) {
	SyncBlock()
	d, _ := RecoverAddress("0xf42400x647b226d657461223a222f697066732f516d52357756784a467574685577626d77585a5570394a6b5069314c5051644a596a716834754a77446d6d5a6758222c22746f6b656e5f6964223a2235323036333831343038313133227d10xa1e67a33e090afe696d7317e05c506d7687bb2e50xae24", "0x6170e56c83bb2324d608af2036dd7cdb329982d4a1cb38e2b3221625af4896c8233619c8cbb6f8f3f5b2c1a13a7f21b942e40658d31c5713a43459382b9007151b")
	fmt.Println(d.String())
}

func Sign(data common.Hash) (string, error) {
	key, err := crypto.HexToECDSA("bf4592f5ca28531bafab976f382963c10a49d765f5d357af07ca83e0bf3d93b7")
	if err != nil {
		return "", err
	}
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n32%s", data.Bytes())
	sig, err := crypto.Sign(crypto.Keccak256([]byte(msg)), key)
	if err != nil {
		return "", err
	}
	sig[64] += 27
	return hexutil.Encode(sig), err
}
func toBytes(v string) []byte {
	var bigTemp big.Int
	bigTemp.SetString(v, 0)
	return bigTemp.Bytes()
}
func GenStakingNftSign1() (error, string) {
	data := crypto.Keccak256Hash(
		common.LeftPadBytes(toBytes("0xa"), 32),
		common.HexToAddress("0x8000000000000000000000000000000000000002").Bytes(),
		common.HexToAddress("0x5eca602ab2912ea95835f93555a7eefaa21246dd").Bytes(),
		common.LeftPadBytes(toBytes("0x40520"), 32),
	)

	sign, err := Sign(data)
	if err != nil {

		return err, ""
	}
	fmt.Println(sign)
	return nil, sign
}
func HashMsg(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), string(data))
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	return hasher.Sum(nil), msg
}

// recoverAddress recover the address from sig
func RecoverAddress(msg string, sigStr string) (common.Address, error) {
	sigData := hexutil.MustDecode(sigStr)
	if len(sigData) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sigData[64] != 27 && sigData[64] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sigData[64] -= 27
	hash, _ := HashMsg([]byte(msg))
	rpk, err := crypto.SigToPub(hash, sigData)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}
