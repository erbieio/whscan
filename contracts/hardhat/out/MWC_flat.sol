// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;



contract MWC is ERC20 {
    constructor(address _exchange, address _release, address _firstRelease, address _nft) ERC20("MWC", "MWC"){
        _mint(_exchange, 40 * 1e26);
        _mint(_release, 37 * 1e26);
        _mint(_firstRelease, 18 * 1e26);
        _mint(_nft, 5 * 1e26);
    }
}



