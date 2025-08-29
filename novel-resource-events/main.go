package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"novel-resource-events/chaincode"
)

func main() {
	novelContract := new(chaincode.SmartContract)
	chaincode, err := contractapi.NewChaincode(novelContract)
	if err != nil {
		log.Panicf("Error creating novel chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting novel chaincode: %v", err)
	}
}