Blockchain browser interface service
-------------------------------------------------------------
Obtain and analyze the block transactions of the specified blockchain nodes, etc., and provide an interface for querying and analyzing data

## Blockchain Network
The predefined blockchain network is in the file `conf/networks.go`, mainly chain ID and node RPC link, if you need a special blockchain network, you can modify this file
```
var networks = map[int64]*network{
	1337: {
		Name: "localhost",
		Url:  "http://127.0.0.1:8545",
	},
	......
	51888: {
		Name: "wormholes",
		Url:  "http://43.129.181.130:8561",
	},
	51889: {
		Name: "wormholes dev",
		Url:  "https://api.wormholestest.com",
	},
	80001: {
		Name: "mumbai",
		Url:  "https://rpc-mumbai.maticvigil.com",
	},
}
```


## configure
Create a file `.env` (same execution path as the executable file), the following is the default configuration, fill in whatever fields need to be modified

```
CHAIN_ID    =51888
HEX_KEY     =7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd
SERVER_ADDR =:3000
INTERVAL    =10
MYSQL_DSN   =root:123456@tcp(127.0.0.1:3306)/scan
RESET_DB    =true
IPFS_SERVER =http://localhost:8080
AMOUNT_STR  =100000000000000000000
ERB_PAY     =0xa03196bF28ffABcab352fe6d58F4AA83998bebA1
```

1. CHAIN_ID: Specifies the chain ID blockchain data to be analyzed, see the predefined blockchain network file for details
2. HEX_KEY: hexadecimal account private key without 0x prefix (length 64 bits), send test currency account
3. SERVER_ADDR: Open query service interface address, after running, query and analyze data through this address
4. INTERVAL: The pause time (in seconds) when there is an error in the analysis or when there is no new block to analyze
5. MYSQL_DSN: The connection address of the database (mysql or mariadb database)
6. RESET_DB: Automatically clear the database when the operation starts
7. IPFS_SERVER: IPFS server address, used for parsing NFT meta information
8. AMOUNT_STR: The amount of coins sent by the faucet at one time
9. ERB_PAY: ERB Pay contract address

## Dedicated blockchain node
The node parameters to start must contain at least:
1. Enable http (ws) service: `--http` (`--ws`)
2. Open query api (minimum eth and debug): `--http.api debug,eth` (`--ws.api debug,eth`)
3. Select full data mode for synchronization: `--syncmode full`
4. Select archive mode for garbage collection: `--gcmode archive`
5. Keep all transactions: `--txlookuplimit 0`

Example:
```
geth --rinkeby --http --http.api debug,eth --syncmode full --txlookuplimit 0 --gcmode archive
```

## run
1. Configure the predefined blockchain network information (if required), compile the program
2. Install and configure mysql or mariadb database
3. Run the blockchain service node (if needed)
4. Create a configuration file (if needed)
5. Clear the database when the chain resets or switches
6. To run the program, the configuration file needs to be in the same execution path as the program