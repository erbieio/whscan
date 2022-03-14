// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

enum Level {
    None,
    Miner, //可以发行NFT
    Super //可以转移任意帐户的NFT
}

// 授权接口
interface IApproval {
    // 授权更新事件
    event ApprovalLevel(address indexed addr, Level level);

    // 设置权限
    function setLevel(address _addr, Level _level) external;

    // 判断是不是miner身份
    function isMiner(address _miner) external view returns (bool);

    // 判断是不是super身份，super身份包含miner身份
    function isSuper(address _super) external view returns (bool);
}
