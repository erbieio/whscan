package conf

type network struct {
	Name string
	Url  string
}

const InfuraId = "460f40a260564ac4a4f4b3fffb032dad"

var networks = map[int64]*network{
	1337: {
		Name: "localhost",
		Url:  "http://127.0.0.1:8545",
	},
	1: {
		Name: "mainnet",
		Url:  "https://mainnet.infura.io/v3/" + InfuraId,
	},
	3: {
		Name: "ropsten",
		Url:  "https://ropsten.infura.io/v3/" + InfuraId,
	},
	4: {
		Name: "rinkeby",
		Url:  "https://rinkeby.infura.io/v3/" + InfuraId,
	},
	5: {
		Name: "goerli",
		Url:  "https://goerli.infura.io/v3/" + InfuraId,
	},
	42: {
		Name: "kovan",
		Url:  "https://kovan.infura.io/v3/" + InfuraId,
	},
	56: {
		Name: "bsc",
		Url:  "https://bsc-dataseed1.binance.org",
	},
	137: {
		Name: "matic",
		Url:  "https://rpc-mainnet.maticvigil.com",
	},
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
