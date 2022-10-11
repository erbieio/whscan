package conf

import (
	"log"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/joho/godotenv"
	"server/common/utils"
)

// default allocation
var (
	ChainUrl   = "http://localhost:8545"
	HexKey     = "7b2546a5d4e658d079c6b2755c6d7495edd01a686fddae010830e9c93b23e398"
	ServerAddr = ":3000"
	Interval   = time.Second
	Thread     = int64(runtime.NumCPU())
	MysqlDsn   = "root:123456@tcp(127.0.0.1:3306)/scan"
	ResetDB    = false
	IpfsServer = "http://localhost:8080"
	AmountStr  = "1000000000000000000"
	ERBPay     = "0xa03196bF28ffABcab352fe6d58F4AA83998bebA1" //ERBPay contract address
)

// globally available object instantiated from config
var (
	PrivateKey *secp256k1.PrivateKey //Private key
	Amount     *big.Int              //Amount of coins (unit: wei)
)

func init() {
	// set log printout to stdout instead of stderr
	log.SetOutput(os.Stdout)

	// read configuration to override default value
	setConf()

	var err error

	// Blockchain account private key and RPC client configuration
	PrivateKey, err = utils.HexToECDSA(HexKey)
	if err != nil {
		panic(err)
	}
	Amount = new(big.Int)
	Amount.SetString(AmountStr, 0)
}

func setConf() {
	err := godotenv.Load("scan.env")
	if err != nil {
		log.Println("Failed to load environment variables from .env file,", err)
	}

	// Parse the basic configuration of the server
	if chainUrl := os.Getenv("CHAIN_URL"); chainUrl != "" {
		ChainUrl = chainUrl
	}
	if hexKey := os.Getenv("HEX_KEY"); hexKey != "" {
		HexKey = hexKey
	}
	if serverAddr := os.Getenv("SERVER_ADDR"); serverAddr != "" {
		ServerAddr = serverAddr
	}
	if interval := os.Getenv("INTERVAL"); interval != "" {
		Interval, err = time.ParseDuration(interval)
		if err != nil {
			panic(err)
		}
	}
	if thread := os.Getenv("THREAD"); thread != "" {
		Thread, err = strconv.ParseInt(thread, 0, 64)
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
