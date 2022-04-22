package service

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"server/common/utils"
	"server/conf"
	"server/model"
)

const ZeroAddress = "0x0000000000000000000000000000000000000000"

// ErrRes 接口错误信息返回
type ErrRes struct {
	ErrStr string `json:"err_str"` //错误信息
}

func parseLog(l *model.Log) (log utils.Log) {
	log.Data = hexutil.MustDecode(l.Data)
	for _, topic := range l.Topics {
		log.Topics = append(log.Topics, common.HexToHash(topic))
	}
	return
}

// BigIntAdd 两个大数字符串相加,返回十进制大数字符串
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
		panic(nil)
	}
	return cc.Text(10)
}

// NFTMeta NFT核心元信息，只解析这些字段，多余的字段忽略
type NFTMeta struct {
	Name                 string `json:"name"`                  //名称
	Desc                 string `json:"desc"`                  //描述
	Category             string `json:"category"`              //分类
	SourceUrl            string `json:"source_url"`            //资源链接，图片或视频等文件链接
	CollectionsCreator   string `json:"collections_creator"`   //所属合集创建者，唯一标识合集
	CollectionsName      string `json:"collections_name"`      //所属合集的名称，唯一标识合集
	CollectionsCategory  string `json:"collections_category"`  //所属合集的分类
	CollectionsDesc      string `json:"collections_desc"`      //所属合集描述
	CollectionsImgUrl    string `json:"collections_img_url"`   //合集图片链接
	CollectionsExchanger string `json:"collections_exchanger"` //所属合集交易所，唯一标识合集
}

// GetNFTMeta 从链接里获取NFT元信息
func GetNFTMeta(url string) (nft NFTMeta, err error) {
	// 如果ipfs的链接没有给服务器地址，则使用本地ipfs服务器
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
