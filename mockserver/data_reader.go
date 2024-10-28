package mockserver

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type BlockHeader struct {
	BlockHash     chainhash.Hash `json:"hash"`
	Confirmations uint32         `json:"confirmations"`
	Height        int64          `json:"height"`
	Version       int32          `json:"version"`
	VersionHex    string         `json:"versionHex"`
	MerkleRoot    chainhash.Hash `json:"merkleroot"`
	Time          uint32         `json:"time"`
	MedianTime    uint32         `json:"mediantime"`
	Nonce         uint32         `json:"nonce"`
	Bits          string         `json:"bits"`
	Difficulty    uint32         `json:"difficulty"`
	ChainWork     chainhash.Hash `json:"chainwork"`
	NumberTx      uint32         `json:"nTx"`

	PrevBlock chainhash.Hash `json:"previousblockhash"`
	NextBlock chainhash.Hash `json:"nextblockhash"`
}

type Transaction struct {
	TxId          chainhash.Hash `json:"txid"`
	Hash          chainhash.Hash `json:"hash"`
	Version       int32          `json:"version"`
	Size          int32          `json:"size"`
	VSize         int32          `json:"vsize"`
	Weight        int32          `json:"weight"`
	LockTime      uint32         `json:"locktime"`
	VIn           []Vin          `json:"vin"`
	VOut          []btcjson.Vout `json:"vout"`
	Hex           string         `json:"hex"`
	BlockHash     chainhash.Hash `json:"blockhash"`
	Confirmations uint32         `json:"confirmations"`
	Time          uint32         `json:"time"`
	BlockTime     uint32         `json:"blocktime"`
}

// copied from btcjson/chainsvrresults.go and
// modified to be compatible with bitcoin v1
type Vin struct {
	Coinbase string `json:"coinbase"`
	// Txid      string     `json:"txid"`
	// Vout      uint32     `json:"vout"`
	// ScriptSig *ScriptSig `json:"scriptSig"`
	Sequence uint32 `json:"sequence"`
	// Witness  []string `json:"txinwitness"`
}

type DataContent struct {
	BlockHeaders []BlockHeader `json:"block_headers"`
	Transactions []Transaction `json:"transactions"`
}

type DataStore struct {
	DataContent DataContent
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

	d.DataContent = dataContent
}
