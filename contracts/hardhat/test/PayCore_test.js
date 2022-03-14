const {ethers} = require('hardhat')
const {expect} = require('chai')

describe('PayCore', function () {
    it('Should deploy PayCore Contract', async function () {
        const [{address}, , other] = await ethers.getSigners()
        const Titan = await ethers.getContractFactory('Titan')
        const titan = await Titan.deploy("Titan", "tt")
        await titan.deployed()

        const PayCore = await ethers.getContractFactory('PayCore')
        const payCore = await PayCore.deploy(titan.address)
        await payCore.deployed()

        describe('feeInfo', function () {
            it('Should get fee info', async function () {
                const fee = await payCore.feeInfo(1000)
                expect(fee[0]).to.equal(address)
                expect(fee[1].toNumber()).to.equal(30)

                describe('setFeeRatio', function () {
                    it('Should set new fee ratio', async function () {
                        // 将费率设为5%（费率单位是万分之一）
                        await payCore.setFeeRatio(500)
                        const fee = await payCore.feeInfo(100)
                        expect(fee[1].toNumber()).to.equal(5)
                        try {
                            // 非合约所有者调用将会revert并抛出错误
                            await payCore.connect(other).setFeeRatio(500)
                            assert(false)
                        } catch (error) {
                            expect(error.message).to.have.string('reverted')
                        }
                    })
                })

                describe('setFeeReciver', function () {
                    it('Should set new fee reciver', async function () {
                        const [, {address}] = await ethers.getSigners()
                        // 更改手续费接收地址
                        await payCore.setFeeReciver(address)
                        const fee = await payCore.feeInfo(100)
                        expect(fee[0]).to.equal(address)
                        try {
                            // 非合约所有者调用将会revert并抛出错误
                            await payCore.connect(other).setFeeReciver(address)
                            assert(false)
                        } catch (error) {
                            expect(error.message).to.have.string('reverted')
                        }
                    })
                })
            })
        })
        describe('payTitan', function () {
            it('payTitan test', async function () {
                // try {
                const addr1 = "0x5FD6eB55D12E759a21C09eF703fe0CBa1DC9d88D"
                const addr2 = "0x10cec672c6bb2f6782beed65987e020902b7bd15"
                await payCore.setFeeReciver(addr1)
                const titanAddress = await payCore.titan()
                console.log("titan: ", titanAddress, titan.address)
                await titan.mint(address, 100000000)
                const b1 = await titan.balanceOf(address)
                const b2 = await titan.balanceOf(addr1)
                console.log("before payTitan ", b1, b2)
                await titan.approve(payCore.address, 10000000000);
                const fee = await payCore.feeInfo(1000)
                console.log("payTitan feeReceiver and  fee ", fee[0], fee[1])
                const [owner] = await ethers.getSigners()
                const res = await payCore.payTitan(address, addr2, 10000)
                const receipt = await res.wait()
                console.log(receipt.logs)
                const b3 = await titan.balanceOf(address)
                const b4 = await titan.balanceOf(addr1)
                const b5 = await titan.balanceOf(addr2)
                console.log("after payTitan ", b3, b4, b5)
                // } catch (error) {
                //     expect(error.message).to.have.string('reverted')
                // }
            })
        })

        describe('titan', function () {
            it('Should setTitan address', async function () {
                const [, {address}] = await ethers.getSigners()
                // 更改Titan地址
                await payCore.setTitan(address)
                const titan = await payCore.titan()
                expect(titan).to.equal(address)
                // try {
                // const [, other] = await ethers.getSigners()
                    // 非合约所有者调用将会revert并抛出错误
                    // await payCore.connect(other).setTitan(address)
                    // assert(false)
                // } catch (error) {
                //     expect(error.message).to.have.string('reverted')
                // }
            })
        })
    })
})