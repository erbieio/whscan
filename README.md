## 默认配置

文件`conf/conf.go`
```
	ChainId    int64 = 80001
	HexKey           = "7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd"
	ServerAddr       = ":3000"
	Interval   int64 = 30
	SaveTime   int64 = 10 * 60
	MysqlDsn         = "root:123456@tcp(127.0.0.1:3306)/chain?charset=utf8mb4&parseTime=True&loc=Local"
```

## 覆盖默认配置

创建文件`.env`(和可执行文件同一执行路径)
需要覆盖什么配置就填什么字段

```
CHAIN_ID    =80001
HEX_KEY     =7bbfec284ee43e328438d46ec803863c8e1367ab46072f7864c07e0a03ba61fd
SERVER_ADDR =:3000
INTERVAL    =60
SAVE_TIME   =600
MYSQL_DSN   =root:123456@tcp(127.0.0.1:3306)/chain?charset=utf8mb4&parseTime=True&loc=Local
```
