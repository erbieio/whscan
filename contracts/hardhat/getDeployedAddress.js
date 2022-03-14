let fs = require('fs');
const contractJson = require('./deployments/bsctest/ExchangeProxy.json');
const Web3 = require('web3');


(async () => {
    let args = process.argv.splice(2);
    let web3;
    if (args[0] === "bsctest") {
        web3 = new Web3(new Web3.providers.HttpProvider("https://bsc.getblock.io/testnet/?api_key=fcf5c609-2f4c-4131-afbb-2e044f051b9c"));
    } else if (args[0] === "bscmainnet") {
        web3 = new Web3(new Web3.providers.HttpProvider("https://bsc.getblock.io/testnet/?api_key=fcf5c609-2f4c-4131-afbb-2e044f051b9c"));
    } else {
        console.log("请输入获取的网络! 如bsctest 或者 bscmainnet")
        return
    }

    const contract = new web3.eth.Contract(contractJson.abi, contractJson.address);
    let output = ""
    const mwc = await contract.methods.mwc().call()
    output += "mwc address :" + mwc + "\n"
    const mwf = await contract.methods.mwf().call()
    output += "mwf address :" + mwf + "\n"
    const mwm = await contract.methods.mwm().call()
    output += "mwm address :" + mwm + "\n"
    const release = await contract.methods.release().call()
    output += "release address :" + release + "\n"
    output += "exchange address :" + contractJson.address + "\n"
    fs.writeFile("./contractAddress.txt", output, function (err) {
    })
    console.log(" success! output to ./contractAddress.txt ")
})()
