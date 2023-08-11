Blockchain browser interface service
-------------------------------------------------------------
Obtain and analyze the block transactions of the specified blockchain nodes, etc., and provide an interface for querying and analyzing data


## configure
Create a file `.env` (same execution path as the executable file), the following is the default configuration, fill in whatever fields need to be modified

```
CHAIN_URL   =http://localhost:8545
SERVER_ADDR =:3000
INTERVAL    =1s
THREAD      =8
MYSQL_DSN   =root:123456@tcp(127.0.0.1:3306)/scan
```

1. CHAIN_URL: Specifies the chain api address blockchain data to be analyzed
2. SERVER_ADDR: Open query service interface address, after running, query and analyze data through this address
3. INTERVAL: The pause time (in seconds) when there is an error in the analysis or when there is no new block to analyze
4. THREAD: Number of parsing coroutines in parallel
5. MYSQL_DSN: The connection address of the database (mysql or mariadb database)

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