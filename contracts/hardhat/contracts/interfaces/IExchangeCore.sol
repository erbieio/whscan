// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IExchangeCore {
    function payBelly(address _from, uint256 _value) external returns (bool);
    function payTitan(address _from, uint256 _value) external returns (bool);
    function cancel(uint256 _tokenId) external ;
}