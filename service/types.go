package service

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"server/conf"
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
		log.Printf("error: the result is negative number, %s+%s=%s", a, b, cc.String())
		//panic("big add err:" + cc.String())
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
	big.NewFloat(143000000000000000),
	big.NewFloat(271000000000000000),
	big.NewFloat(650000000000000000),
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

// NFTMeta NFT core meta information, only these fields are parsed, the extra fields are ignored
type NFTMeta struct {
	Name                 string `json:"name"`                  //name
	Desc                 string `json:"desc"`                  //description
	Attributes           string `json:"attributes"`            //Attributes
	Category             string `json:"category"`              //category
	SourceUrl            string `json:"source_url"`            //Resource links, file links such as pictures or videos
	CollectionsCreator   string `json:"collections_creator"`   //The creator of the collection, uniquely identifies the collection
	CollectionsName      string `json:"collections_name"`      //The name of the collection to which it belongs, uniquely identifying the collection
	CollectionsCategory  string `json:"collections_category"`  //The category of the collection it belongs to
	CollectionsDesc      string `json:"collections_desc"`      //Description of the collection it belongs to
	CollectionsImgUrl    string `json:"collections_img_url"`   //collection image link
	CollectionsExchanger string `json:"collections_exchanger"` //The collection exchange to which it belongs, uniquely identifies the collection
}

// GetNFTMeta gets NFT meta information from the link
func GetNFTMeta(url string) (nft NFTMeta, err error) {
	// If the ipfs link does not give the server address, use the local ipfs server
	realUrl := url
	if strings.Index(url, "/ipfs/Qm") == 0 {
		realUrl = conf.IpfsServer + url
	}

	resp, err := http.Get(realUrl)
	if err != nil {
		return
	}
	data, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	err = json.Unmarshal(data, &nft)
	return
}
