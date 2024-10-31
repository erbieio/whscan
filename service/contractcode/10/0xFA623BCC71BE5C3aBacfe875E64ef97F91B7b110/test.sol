// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

/**
 * @title Owner
 * @dev Set & change owner
 */
contract TestVerify {
 uint256 private storedData;  
  
    // This function sets the value of storedData  
    function set(uint256 x) public {  
        storedData = x;  
    }  
  
    // This function gets the value of storedData  
    function get() public view returns (uint256) {  
        return storedData;  
    }  
} 
