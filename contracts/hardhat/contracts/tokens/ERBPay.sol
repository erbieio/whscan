// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
import "@openzeppelin/contracts/access/Ownable.sol";

contract ERBPay is Ownable{
    uint256 yearFee=200;
    uint256 otherFee=100;
    address public yearFeeAddress;
    address public otherFeeAddress;
    mapping(address=>uint256)public endTime;
    event PayForOpenExchange(address indexed sender, uint256 amount);
    event PayForOpenRenew(address indexed sender, uint256 amount);
    constructor () {
    }

    function setYearFee(uint256 amount) external onlyOwner{
        yearFee=amount;
    }

    function setOtherFee(uint256 amount) external onlyOwner{
        otherFee=amount;
    }

    function setYearFeeAddress(address _addr) external onlyOwner{
        yearFeeAddress=_addr;
    }

    function setOtherFeeAddress(address _addr) external onlyOwner{
        otherFeeAddress=_addr;
    }

    function payForOpenExchange()external payable{
       require(msg.value == yearFee+otherFee ,"msg.value invalid");
       require(endTime[msg.sender]==0 ,"already opened");
       payable(otherFeeAddress).transfer(otherFee);
       payable(yearFeeAddress).transfer(yearFee);
       endTime[msg.sender]=block.timestamp+86400*365;
       emit PayForOpenExchange(msg.sender,msg.value);
    }

    function payForOpenRenew()external payable{
       require(msg.value == yearFee,"msg.value invalid");
       require(endTime[msg.sender] !=0 ,"no exchange exist");
       payable(yearFeeAddress).transfer(yearFee);
       endTime[msg.sender]=endTime[msg.sender]+86400*365;
       emit PayForOpenRenew(msg.sender,msg.value);
    }

    //0 no exchange 1 no fee  2 normal
    function checkAuth(address addr)view external returns(uint256){
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