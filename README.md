## 配置
创建文件`.env`(和可执行文件同一执行路径)
以下是默认配置，需要覆盖什么配置就填什么字段

```
CHAIN_ID    =51888
HEX_KEY     =7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd
SERVER_ADDR =:3000
INTERVAL    =10
MYSQL_DSN   =root:123456@tcp(127.0.0.1:3306)/scan
```

## 专用区块链节点
启动的节点参数至少需要包含：
1. 开启http（ws）服务：`--http`（`--ws`）
2. 开放查询api(最少eth和debug)：`--http.api debug,eth`（`--ws.api debug,eth`）
3. 同步选择全数据模式：`--syncmode full`
4. 垃圾回收选择归档模式：`--gcmode archive`
5. 保留所有交易：`--txlookuplimit 0`

```
geth --rinkeby --http --http.api debug,eth --syncmode full --txlookuplimit 0 --gcmode archive
```