package mockserver

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/btcsuite/btcd/btcjson"
)

type DataContent struct {
	BlockHeaders []btcjson.GetBlockHeaderVerboseResult `json:"block_headers"`
	Transactions []btcjson.TxRawResult                 `json:"transactions"`
}

type DataStore struct {
	DataContent DataContent

	BlockHeaderMap map[int32]btcjson.GetBlockHeaderVerboseResult
	TransactionMap map[string]btcjson.TxRawResult
}

func (d *DataStore) ReadJson(jsonFilePath string) {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer jsonFile.Close()

	// Read the file contents
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Unmarshal the JSON data into the struct
	var dataContent DataContent
	if err := json.Unmarshal(byteValue, &dataContent); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// populate the BlockHeaderMap from dataContent
	d.BlockHeaderMap = make(map[int32]btcjson.GetBlockHeaderVerboseResult)
	for _, blockHeader := range dataContent.BlockHeaders {
		d.BlockHeaderMap[blockHeader.Height] = blockHeader
	}

	// populate the TransactionMap from dataContent
	d.TransactionMap = make(map[string]btcjson.TxRawResult)
	for _, transaction := range dataContent.Transactions {
		d.TransactionMap[transaction.Txid] = transaction
	}

	d.DataContent = dataContent
}
