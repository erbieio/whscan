package service

import (
	"gorm.io/gorm"
	"server/common/model"
)

// ContractsRes contract paging return parameters
type ContractsRes struct {
	Total     int64             `json:"total"`     //The total number of contracts
	Contracts []*model.Contract `json:"contracts"` //contract list
}

func FetchContracts(contractType int, page, size int) (res ContractsRes, err error) {
	var db *gorm.DB
	if contractType == 0 {
		db = DB.Model(&model.Contract{}).Where("contract_type = ?", "ERC20").Order("block_number DESC")
		err = DB.Model(&model.Contract{}).Where("contract_type = ?", "ERC20").Count(&res.Total).Error
		if err != nil {
			return
		}
	} else if contractType == 1 {
		db = DB.Model(&model.Contract{}).Where("contract_type = ? || contract_type = ?", "ERC721", "ERC1155").Order("block_number DESC")
		err = DB.Model(&model.Contract{}).Where("contract_type = ? || contract_type = ?", "ERC721", "ERC1155").Count(&res.Total).Error
		if err != nil {
			return
		}
	}

	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Contracts).Error
	if err != nil {
		return
	}

	for idx, contract := range res.Contracts {
		var holders int64
		if contract.ContractType == "ERC20" {
			err = DB.Model(&model.ContractAccountErc20{}).Where("contract_address = ?", contract.ContractAddress).Count(&holders).Error
			if err != nil {
				return
			}
		}
		if contract.ContractType == "ERC721" ||
			contract.ContractType == "ERC1155" {
			err = DB.Model(&model.ContractNFT{}).Where("contract_address = ?", contract.ContractAddress).Count(&holders).Error
			if err != nil {
				return
			}
			var tmpNft model.ContractNFT
			err = DB.Model(&model.ContractNFT{}).Where("contract_address = ?", contract.ContractAddress).Order("token_id desc").First(&tmpNft).Error
			if err != nil {
				return
			}
			res.Contracts[idx].TotalSupply = string(tmpNft.TokenId)
		}
		res.Contracts[idx].Holders = holders
	}
	return
}

func GetContract(addr string) (*model.Contract, error) {
	var contract model.Contract
	err := DB.Model(&model.Contract{}).Where("contract_address = ?", addr).First(&contract).Error
	if err != nil {
		return nil, err
	}

	var holders int64
	if contract.ContractType == "ERC20" {
		err = DB.Model(&model.ContractAccountErc20{}).Where("contract_address = ?", contract.ContractAddress).Count(&holders).Error
		if err != nil {
			return nil, err
		}
	}
	if contract.ContractType == "ERC721" ||
		contract.ContractType == "ERC1155" {
		err = DB.Model(&model.ContractNFT{}).Where("contract_address = ?", contract.ContractAddress).Count(&holders).Error
		if err != nil {
			return nil, err
		}
		var tmpNft model.ContractNFT
		err = DB.Model(&model.ContractNFT{}).Where("contract_address = ?", contract.ContractAddress).Order("token_id desc").First(&tmpNft).Error
		if err != nil {
			return nil, err
		}
		contract.TotalSupply = string(tmpNft.TokenId)
	}
	contract.Holders = holders

	return &contract, nil
}

// HoldersRes contract holders paging return parameters
type HolderInfo struct {
	Address  string  `json:"address"`
	Quantity float64 `json:"quantity"`
}
type HoldersRes struct {
	TotalCount  int64         `json:"total_count"` //The total number of holders
	TotalAmount string        `json:"total_amount"`
	Holders     []*HolderInfo `json:"holders"` //holders list
}

func FetchHolders(addr string, page, size int) (*HoldersRes, error) {
	var holdersRes HoldersRes
	var contract model.Contract
	err := DB.Model(&model.Contract{}).Where("contract_address = ?", addr).First(&contract).Error
	if err != nil {
		return nil, err
	}
	if contract.ContractType == "ERC20" {
		holdersRes.TotalAmount = contract.TotalSupply
		err = DB.Model(&model.ContractAccountErc20{}).Where("contract_address = ?", addr).Count(&holdersRes.TotalCount).Error
		if err != nil {
			return nil, err
		}

		err = DB.Model(&model.ContractAccountErc20{}).Select("address", "balance").Where("contract_address = ?",
			addr).Order("balance DESC").Offset((page - 1) * size).Limit(size).Find(&holdersRes.Holders).Error
		if err != nil {
			return nil, err
		}
	} else if contract.ContractType == "ERC721" ||
		contract.ContractType == "ERC1155" {
		var tmpNft model.ContractNFT
		err = DB.Model(&model.ContractNFT{}).Where("contract_address = ?", addr).Order("token_id desc").First(&tmpNft).Error
		if err != nil {
			return nil, err
		}
		holdersRes.TotalAmount = string(tmpNft.TokenId)
		err = DB.Model(&model.ContractNFT{}).Select("owner as address",
			"count(owner) as quantity").Group("owner").Count(&holdersRes.TotalCount).Error
		if err != nil {
			return nil, err
		}

		err = DB.Model(&model.ContractNFT{}).Select("owner as address",
			"count(owner) as quantity").Group("owner").Order("quantity desc").Offset((page - 1) * size).Limit(size).Find(&holdersRes.Holders).Error
		if err != nil {
			return nil, err
		}
	}

	return &holdersRes, nil
}
