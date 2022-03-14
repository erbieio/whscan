require("@nomiclabs/hardhat-ethers");
require("hardhat-deploy");

const fs = require("fs");
const defaultNetwork = "localhost";

function mnemonic() {
    try {
        return fs.readFileSync("./mnemonic.txt").toString().trim();
    } catch (e) {
        if (defaultNetwork !== "localhost") {
            console.log(
                "☢️ WARNING: No mnemonic file created for a deploy account. Try `yarn run generate` and then `yarn run account`."
            );
        }
    }
    return "";
}

module.exports = {
    defaultNetwork,

    // don't forget to set your provider like:
    // REACT_APP_PROVIDER=https://dai.poa.network in packages/react-app/.env
    // (then your frontend will talk to your contracts on the live network!)
    // (you will need to restart the `yarn run start` dev server after editing the .env)

    networks: {
        hardhat: {
            chainId: 1337
        },
        localhost: {

            /*
              notice no mnemonic here? it will just use account 0 of the hardhat node to deploy
              (you can put in a mnemonic here to set the deployer locally)
            */
        },
        bsctest: {
            url: "https://bsc.getblock.io/testnet/?api_key=fcf5c609-2f4c-4131-afbb-2e044f051b9c", // <---- YOUR INFURA ID! (or it won't work)
            gasPrice: 1800000010,
            accounts: {
                mnemonic: mnemonic(),
            },
        },
        bscmainnet: {
            url: "https://bsc.getblock.io/mainnet/?api_key=fcf5c609-2f4c-4131-afbb-2e044f051b9c", // <---- YOUR INFURA ID! (or it won't work)
            gasPrice: 1800000010,
            accounts: {
                mnemonic: mnemonic(),
            },
        },
        rinkeby: {
            url: "https://rinkeby.infura.io/v3/460f40a260564ac4a4f4b3fffb032dad", // <---- YOUR INFURA ID! (or it won't work)
            gasPrice: 1100000010,

            accounts: {
                mnemonic: mnemonic(),
            },
        },
        kovan: {
            url: "https://kovan.infura.io/v3/460f40a260564ac4a4f4b3fffb032dad", // <---- YOUR INFURA ID! (or it won't work)
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
        },
        mainnet: {
            url: "https://mainnet.infura.io/v3/460f40a260564ac4a4f4b3fffb032dad", // <---- YOUR INFURA ID! (or it won't work)
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
        },
        ropsten: {
            url: "https://ropsten.infura.io/v3/460f40a260564ac4a4f4b3fffb032dad", // <---- YOUR INFURA ID! (or it won't work)
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
        },
        goerli: {
            gasPrice: 28180409118,
            url: "https://goerli.infura.io/v3/460f40a260564ac4a4f4b3fffb032dad", // <---- YOUR INFURA ID! (or it won't work)
            accounts: {
                mnemonic: mnemonic(),
            },
        },
        xdai: {
            url: "https://rpc.xdaichain.com/",
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
        },
        matic: {
            url: "https://rpc-mainnet.maticvigil.com/",
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
        },
        rinkebyArbitrum: {
            url: "https://rinkeby.arbitrum.io/rpc",
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
            companionNetworks: {
                l1: "rinkeby",
            },
        },
        localArbitrum: {
            url: "http://localhost:8547",
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
            companionNetworks: {
                l1: "localArbitrumL1",
            },
        },
        localArbitrumL1: {
            url: "http://localhost:7545",
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
            companionNetworks: {
                l2: "localArbitrum",
            },
        },
        kovanOptimism: {
            url: "https://kovan.optimism.io",
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
            ovm: true,
            companionNetworks: {
                l1: "kovan",
            },
        },
        localOptimism: {
            url: "http://localhost:8545",
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
            ovm: true,
            companionNetworks: {
                l1: "localOptimismL1",
            },
        },
        localOptimismL1: {
            url: "http://localhost:9545",
            gasPrice: 28180409118,
            accounts: {
                mnemonic: mnemonic(),
            },
            companionNetworks: {
                l2: "localOptimism",
            },
        },
    },
    solidity: {
        compilers: [
            {
                version: "0.8.4",
                settings: {
                    optimizer: {
                        enabled: true,
                        runs: 200,
                    },
                },
            },
            {
                version: "0.6.7",
                settings: {
                    optimizer: {
                        enabled: true,
                        runs: 200,
                    },
                },
            },
        ],
    },
    namedAccounts: {
        deployer: {
            default: 0, // here this will by default take the first account as deployer
        },
    },
};

task("wallet", "Create a wallet (pk) link", async (_, {ethers}) => {
    const randomWallet = ethers.Wallet.createRandom();
    const privateKey = randomWallet._signingKey().privateKey;
    console.log("🔐 WALLET Generated as " + randomWallet.address + "");
    console.log("🔗 http://localhost:3000/pk#" + privateKey);
});

task("accounts", "Prints the list of accounts", async (_, {ethers}) => {
    const accounts = await ethers.provider.listAccounts();
    accounts.forEach((account) => console.log(account));
});

task("blockNumber", "Prints the block number", async (_, {ethers}) => {
    const blockNumber = await ethers.provider.getBlockNumber();
    console.log(blockNumber);
});

async function addr(ethers, addr, utils) {
    const {isAddress, getAddress} = utils;
    if (isAddress(addr)) {
        return getAddress(addr);
    }
    const accounts = await ethers.provider.listAccounts();
    if (accounts[addr] !== undefined) {
        return accounts[addr];
    }
    throw `不是规范的地址: ${addr}`;
}

task("send", "发送ETH")
    .addParam("from", "发送者地址或者节点上的账户序号")
    .addOptionalParam("to", "接收者地址或者节点上的账户序号")
    .addOptionalParam("amount", "要发送的ETH数量")

    .setAction(async (taskArgs, {ethers}) => {
        const {parseUnits} = ethers.utils;
        const from = await addr(ethers, taskArgs.from, ethers.utils);
        const fromSigner = await ethers.provider.getSigner(from);
        const to = await addr(ethers, taskArgs.to, ethers.utils);
        const value = parseUnits(taskArgs.amount, "ether").toHexString();
        return fromSigner.sendTransaction({to, value})
    });