const conf = require("../config");
module.exports = async ({network, getNamedAccounts, deployments}) => {
    const {deploy} = deployments;
    const {deployer} = await getNamedAccounts();

    const exchangeProxy = await deploy("ExchangeProxy", {
        from: deployer,
        args: [conf.fund, conf.team, conf.firstRelease, conf.nft, conf.market, conf.nature, conf.black, conf.nftFarmer, conf.developer, conf.verify],
        log: true,
    });
};
module.exports.tags = ["exchangeProxy"];
