package conf

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// default allocation
var (
	ChainUrl   = "http://localhost:8545"
	ServerAddr = ":3000"
	Interval   = time.Second
	Thread     = int64(16 * runtime.NumCPU())
	MysqlDsn   = "root:123456@tcp(127.0.0.1:3306)/scan"
	IpfsServer = "http://localhost:8080"
	ERBPay     = "0xa03196bF28ffABcab352fe6d58F4AA83998bebA1" //ERBPay contract address
)

func init() {
	// set log printout to stdout instead of stderr
	log.SetOutput(os.Stdout)

	// read configuration to override default value
	err := godotenv.Load("scan.env")
	if err != nil {
		log.Println("Failed to load environment variables from .env file,", err)
	}

	// Parse the basic configuration of the server
	if chainUrl := os.Getenv("CHAIN_URL"); chainUrl != "" {
		ChainUrl = chainUrl
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
	if ipfsServer := os.Getenv("IPFS_SERVER"); ipfsServer != "" {
		IpfsServer = ipfsServer
	}
	if erbPay := os.Getenv("ERB_PAY"); erbPay != "" {
		ERBPay = erbPay
	}
}
