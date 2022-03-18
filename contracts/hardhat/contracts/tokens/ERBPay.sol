// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
import "@openzeppelin/contracts/access/Ownable.sol";

contract ERBPay is Ownable{
    uint256 yearFee=200*1e18;
    address public yearFeeAddress=0x10CEc672c6BB2f6782BEED65987E020902B7bD15;
    mapping(address=>uint256)public endTime;
    event PayForOpenRenew(address indexed sender, uint256 amount);
    constructor () {
    }

    function setYearFee(uint256 amount) external onlyOwner{
        yearFee=amount;
    }

    function setYearFeeAddress(address _addr) external onlyOwner{
        yearFeeAddress=_addr;
    }

    function payForRenew()external payable{
       require(msg.value == yearFee,"msg.value invalid");
       payable(yearFeeAddress).transfer(yearFee);
       endTime[msg.sender]=block.timestamp+86400*365;
       emit PayForOpenRenew(msg.sender,msg.value);
    }

    //0 no exchange 1 no fee  2 normal
    function checkAuth(address addr)view public returns(uint256){
        if(endTime[addr]==0){
           return 0;
        }else if(endTime[addr]<block.timestamp){
           return 1;
        }
        return 2;
    }
    
    fallback() payable external {}
    receive() payable external {}
}