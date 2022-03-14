// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";

interface IERC721Extension is IERC721{
    function burnBatch(uint256[] memory tokenId) external ;
    function changeNftLock(uint256 _tokenId, bool _status) external;
    function mint(address _to,uint256 _tokenId) external;
    function burn(uint256 _Id)  external;
    function isLock(uint256 _tokenId)view external returns(bool);
}