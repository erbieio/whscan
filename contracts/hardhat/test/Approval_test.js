const { ethers } = require('hardhat')
const { expect, assert } = require('chai')

describe('Approval', () => {
    it('Should deploy Approval Contract', async () => {
        const Approval = await ethers.getContractFactory('Approval')
        const approval = await Approval.deploy()
        await approval.deployed()

        describe('setLevel', () => {
            it('Should set approval level', async () => {
                const [owner, user1, user2, user3, user4] = await ethers.getSigners()
                // 授予Super权限
                await approval.setLevel(user1.address, 2)
                // 授予Miner权限
                await approval.setLevel(user2.address, 1)
                // 关闭权限
                await approval.setLevel(user3.address, 0)
                expect(await approval.owner()).to.equal(owner.address)
                expect(await approval.isSuper(owner.address)).to.equal(true)
                expect(await approval.isSuper(user1.address)).to.equal(true)
                expect(await approval.isMiner(user2.address)).to.equal(true)
                expect(await approval.isMiner(user3.address)).to.equal(false)
                expect(await approval.isMiner(user4.address)).to.equal(false)
                await approval.setLevel(user2.address, 0)
                expect(await approval.isMiner(user2.address)).to.equal(false)
                try {
                    // 非合约所有者调用将会revert并抛出错误
                    await approval.connect(user4).setLevel(user3.address, 2)
                    assert(false)
                } catch (error) {
                    expect(error.message).to.have.string('reverted')
                }
            })
        })
    })
})