## configure
Create the file `.env` (same execution path as the executable)
The following is the default configuration, fill in whatever fields need to be overridden

```
CHAIN_ID    =51888
HEX_KEY     =7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd
SERVER_ADDR =:3000
INTERVAL    =10
MYSQL_DSN   =root:123456@tcp(127.0.0.1:3306)/scan
```

## Dedicated blockchain node
The node parameters to start must contain at least:
1. Enable http (ws) service: `--http` (`--ws`)
2. Open query api (minimum eth and debug): `--http.api debug,eth` (`--ws.api debug,eth`)
3. Select full data mode for synchronization: `--syncmode full`
4. Select archive mode for garbage collection: `--gcmode archive`
5. Keep all transactions: `--txlookuplimit 0`

```
geth --rinkeby --http --http.api debug,eth --syncmode full --txlookuplimit 0 --gcmode archive
```