// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

library Verify {
    function recoverSigner(bytes memory _data, bytes memory _sig) public pure returns (address){
        require(_sig.length == 65);
        (uint8 v, bytes32 r, bytes32 s) = (0, 0, 0);
        assembly {
            r := mload(add(_sig, 32))
            s := mload(add(_sig, 64))
            v := byte(0, mload(add(_sig, 96)))
        }

        return ecrecover(_signHash(_data), v, r, s);
    }

    function _signHash(bytes memory _data) public pure returns (bytes32) {
        return
        keccak256(
            abi.encodePacked(
                "\x19Ethereum Signed Message:\n32",
                keccak256(_data)
            )
        );
    }
}
