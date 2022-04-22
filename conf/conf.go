package conf

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/joho/godotenv"
	"server/common/utils"
)

// 默认配置
var (
	ChainId    int64 = 51888
	HexKey           = "7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd"
	ServerAddr       = ":3000"
	Interval   int64 = 10
	MysqlDsn         = "root:123456@tcp(127.0.0.1:3306)/scan"
	ResetDB          = false
	IpfsServer       = "http://localhost:8080"
	AmountStr        = "100000000000000000000"
	ERBPay           = "0x0E699BFe34B3Fc9DA920607DCB32eeeC53E7f420" //ERBPay合约地址
)

// 从配置实例化的全局可用对象
var (
	ChainUrl   string             //链节点地址
	PrivateKey *ecdsa.PrivateKey  //私钥
	Transactor *bind.TransactOpts //智能合约调用者，由私钥创建
	Amount     *big.Int           //币数量（单位：wei）
)

func init() {
	// 读取配置覆盖默认值
	setConf()

	// 校验配置
	if Interval < 10 {
		panic("conf.Interval < 5")
	}

	var err error
	// 区块链网络配置
	Network := networks[ChainId]
	if Network == nil {
		panic(fmt.Sprintf("不支持的chainId：%v", ChainId))
	}

	// 区块链账户私钥和RPC客户端配置
	PrivateKey, err = utils.HexToECDSA(HexKey)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	Transactor, err = bind.NewKeyedTransactorWithChainID(PrivateKey, big.NewInt(ChainId))
	if err != nil {
		panic(err)
	}
	ChainUrl = Network.Url
	Amount = new(big.Int)
	Amount.SetString(AmountStr, 0)
}

func setConf() {
	err := godotenv.Load("scan.env")
	if err != nil {
		log.Println("从.env文件加载环境变量失败，", err)
	}

	// 解析服务器基本配置
	if chainId := os.Getenv("CHAIN_ID"); chainId != "" {
		ChainId, err = strconv.ParseInt(chainId, 0, 64)
		if err != nil {
			panic(err)
		}
	}
	if hexKey := os.Getenv("HEX_KEY"); hexKey != "" {
		HexKey = hexKey
	}
	if serverAddr := os.Getenv("SERVER_ADDR"); serverAddr != "" {
		ServerAddr = serverAddr
	}
	if interval := os.Getenv("INTERVAL"); interval != "" {
		Interval, err = strconv.ParseInt(interval, 0, 64)
		if err != nil {
			panic(err)
		}
	}
	if mysqlDsn := os.Getenv("MYSQL_DSN"); mysqlDsn != "" {
		MysqlDsn = mysqlDsn
	}
	if resetDB := os.Getenv("RESET_DB"); resetDB != "" {
		ResetDB = resetDB == "true"
	}
	if ipfsServer := os.Getenv("IPFS_SERVER"); ipfsServer != "" {
		IpfsServer = ipfsServer
	}
	if amountStr := os.Getenv("AMOUNT_STR"); amountStr != "" {
		AmountStr = amountStr
	}
	if erbPay := os.Getenv("ERB_PAY"); erbPay != "" {
		ERBPay = erbPay
	}
}
