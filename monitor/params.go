package monitor

import common2 "server/ethhelper/common"

var (
	erc721Input  = "0x01ffc9a7" + common2.Erc721Interface + "00000000000000000000000000000000000000000000000000000000"
	erc1155Input = "0x01ffc9a7" + common2.Erc1155Interface + "00000000000000000000000000000000000000000000000000000000"

	erc721Or20TransferEvent = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	erc1155TransferEvent    = "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62"
	nameHash                = "0x06fdde03"
	symbolHash              = "0x95d89b41"
)
