package service

import (
	"encoding/json"
	"io/ioutil"
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
		panic("big add err:" + a + " " + b)
	}
	return cc.Text(10)
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
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &nft)
	return
}
