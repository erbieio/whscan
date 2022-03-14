// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface ILpMining {
    struct pool {
        uint256 poolId;
        string coinName;
        address rewordCoin;
        address lpCoin;
        uint256 totalRewords;
        uint256 remainTotalRewords;
        uint256 blockRewords;
        uint32 fee;
        uint256 startBlock;
        uint256 totalWeight;
        uint256 lastRewardBlock;
        uint256 accPerShare;
    }

    struct userInfo {
        uint256 totalWeight;
        uint256 rewardDebt;
        uint256 lastDeposit;
        uint256 lockRewards;
    }
}
