package conf

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/joho/godotenv"
	"server/common/utils"
)

// default allocation
var (
	ChainId    int64 = 51888
	HexKey           = "7b2546a5d4e658d079c6b2755c6d7495edd01a686fddae010830e9c93b23e398"
	ServerAddr       = ":3000"
	Interval   int64 = 10
	MysqlDsn         = "root:123456@tcp(127.0.0.1:3306)/scan"
	ResetDB          = false
	IpfsServer       = "http://localhost:8080"
	AmountStr        = "1000000000000000000"
	ERBPay           = "0xa03196bF28ffABcab352fe6d58F4AA83998bebA1" //ERBPay contract address
)

// globally available object instantiated from config
var (
	ChainUrl   string                //Chain node address
	PrivateKey *secp256k1.PrivateKey //Private key
	Amount     *big.Int              //Amount of coins (unit: wei)
)

func init() {
	// set log printout to stdout instead of stderr
	log.SetOutput(os.Stdout)

	// read configuration to override default value
	setConf()

	// check configuration
	if Interval < 10 {
		panic("conf.Interval < 5")
	}

	var err error
	// Blockchain network configuration
	Network := networks[ChainId]
	if Network == nil {
		panic(fmt.Sprintf("Unsupported chainId: %v", ChainId))
	}

	// Blockchain account private key and RPC client configuration
	PrivateKey, err = utils.HexToECDSA(HexKey)
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
		log.Println("Failed to load environment variables from .env file,", err)
	}

	// Parse the basic configuration of the server
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
