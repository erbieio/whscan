// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface ITitanCore {
    function cardTransfer(address _from, address _to, uint256 _tokenId) external ;
    function calculateFighting(address _target) external view returns (uint256);
}
