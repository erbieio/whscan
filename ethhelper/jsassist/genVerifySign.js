const ethers = require("ethers");

(async () => {
    try {
        let args = process.argv.splice(2);
        const privateKey = "0x8c995fd78bddf528bd548cce025f62d4c3c0658362dbfd31b23414cf7ce2e8ed";
        const mintData = ethers.utils.solidityKeccak256(['uint256'],
            [parseInt(args[0])])

        const sig3 = await (new ethers.Wallet(privateKey)).signMessage(ethers.utils.arrayify(mintData));
        console.log("true," + ethers.utils.joinSignature(sig3));
    } catch (e) {
        console.log("true," + e.toString())
    }

})()