package service

import (
	"encoding/json"
	"io"
	"math/big"
	"net/http"
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
		panic("big add err:" + cc.String())
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

func snftValue(snft string, count int64) string {
	b := big.NewInt(count)
	switch 42 - len(snft) {
	case 0:
		return b.Mul(b, big.NewInt(30000000000000000)).Text(10)
	case 1:
		return b.Mul(b, big.NewInt(143000000000000000)).Text(10)
	case 2:
		return b.Mul(b, big.NewInt(271000000000000000)).Text(10)
	default:
		return b.Mul(b, big.NewInt(650000000000000000)).Text(10)
	}
}

func snftMergeValue(snft string, count int64) string {
	b := big.NewInt(count)
	switch 42 - len(snft) {
	case 1:
		return b.Mul(b, big.NewInt(113000000000000000)).Text(10)
	case 2:
		return b.Mul(b, big.NewInt(128000000000000000)).Text(10)
	default:
		return b.Mul(b, big.NewInt(379000000000000000)).Text(10)
	}
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
