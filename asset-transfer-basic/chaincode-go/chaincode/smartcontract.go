package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism accross languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	webfilterlist int    `json:"webfilterlist"`
	blocklist     string `json:"blocklist"`
	allowlist     string `json:"allowlist"`
	attribute1    string `json:"attribute1"`
	attribute2    int    `json:"attribute2"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{allowlist: "www.google.com", blocklist: "", attribute2: 5, attribute1: "", webfilterlist: 300},
		{allowlist: "", blocklist: "www.xxx.com", attribute2: 5, attribute1: "", webfilterlist: 400},
		{allowlist: "www.bbc.co.uk", blocklist: "", attribute2: 10, attribute1: "", webfilterlist: 500},
		{allowlist: "https://scholar.google.com/", blocklist: "", attribute2: 10, attribute1: "", webfilterlist: 600},
		{allowlist: "", blocklist: "www.instagram.com", attribute2: 15, attribute1: "", webfilterlist: 700},
		{allowlist: "www.napier.ac.uk", blocklist: "", attribute2: 15, attribute1: "", webfilterlist: 800},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.allowlist, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, allowlist string, blocklist string, attribute2 int, attribute1 string, webfilterlist int) error {
	exists, err := s.AssetExists(ctx, allowlist)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", allowlist)
	}

	asset := Asset{
		allowlist:     allowlist,
		blocklist:     blocklist,
		attribute2:    attribute2,
		attribute1:    attribute1,
		webfilterlist: webfilterlist,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(allowlist, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given allowlist.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, allowlist string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(allowlist)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", allowlist)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provallowlisted parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, allowlist string, blocklist string, attribute2 int, attribute1 string, webfilterlist int) error {
	exists, err := s.AssetExists(ctx, allowlist)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", allowlist)
	}

	// overwriting original asset with new asset
	asset := Asset{
		allowlist:     allowlist,
		blocklist:     blocklist,
		attribute2:    attribute2,
		attribute1:    attribute1,
		webfilterlist: webfilterlist,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(allowlist, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, allowlist string) error {
	exists, err := s.AssetExists(ctx, allowlist)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", allowlist)
	}

	return ctx.GetStub().DelState(allowlist)
}

// AssetExists returns true when asset with given allowlist exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, allowlist string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(allowlist)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the attribute1 field of asset with given allowlist in world state, and returns the old attribute1.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, allowlist string, newattribute1 string) (string, error) {
	asset, err := s.ReadAsset(ctx, allowlist)
	if err != nil {
		return "", err
	}

	oldattribute1 := asset.attribute1
	asset.attribute1 = newattribute1

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(allowlist, assetJSON)
	if err != nil {
		return "", err
	}

	return oldattribute1, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
