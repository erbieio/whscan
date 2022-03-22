const Web3 = require('web3')
const ERBPay = require('./ERBPay.json')
const Token = require('./Token.json')

//const endpoint = 'https://data-seed-prebsc-1-s2.binance.org:8545/';
const endpoint = 'http://192.168.1.235:8561';
const hexPrivateKey = '0xbf4592f5ca28531bafab976f382963c10a49d765f5d357af07ca83e0bf3d93b7';

async function sendTransaction(web3, chainId, account, data, nonce, gasPrice) {
    const message = {
        from: account.address,
        gas: 5000000,
        gasPrice: gasPrice,
        data: data.startsWith('0x') ? data : '0x' + data,
        nonce: nonce,
        chainId: chainId
    }
    const transaction = await account.signTransaction(message)
    return web3.eth.sendSignedTransaction(transaction.rawTransaction)
}

(async () => {
    const options = { timeout: 1000 * 30 }
    const web3 = new Web3(new Web3.providers.HttpProvider(endpoint, options))
    const account = web3.eth.accounts.privateKeyToAccount(hexPrivateKey)

    const chainId = await web3.eth.getChainId()
    const gasPrice = await web3.eth.getGasPrice()
    let nonce = await web3.eth.getTransactionCount(account.address)


    console.info('*******************************************')
    // deploy ERBPay contract
    let factory = null
    {
        const contract = new web3.eth.Contract(ERBPay.abi)
        const options = { data: ERBPay.bytecode, arguments: [] }
        const data = contract.deploy(options).encodeABI()
        const receipt = await sendTransaction(web3, chainId, account, data, nonce, gasPrice)
        console.info('ERBPay:', factory = receipt.contractAddress)
        nonce = nonce + 1
    }

    // deploy Token contract
    {
        const contract = new web3.eth.Contract(Token.abi)
        const options = { data: Token.bytecode, arguments: [] }
        const data = contract.deploy(options).encodeABI()
        const receipt = await sendTransaction(web3, chainId, account, data, nonce, gasPrice)
        console.info('ERC20 Token:', receipt.contractAddress)
        nonce = nonce + 1
    }
    console.info('*******************************************')
})()
