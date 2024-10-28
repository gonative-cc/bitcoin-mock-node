package mockserver

import (
	"fmt"
	"net/http/httptest"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/filecoin-project/go-jsonrpc"
)

// Have a type with some exported methods
type MockServerHandler struct {
	DataStore DataStore
}

func (h *MockServerHandler) PopulateDataStore() {
	h.DataStore.ReadJson("../data/mainnet_oldest_blocks.json")
}

func (h *MockServerHandler) Ping(in int) int {
	return in
}

func (h *MockServerHandler) GetBestBlockHash() (*chainhash.Hash, error) {
	// find the highest block height
	maxHeightIndex := 0
	maxHeight := h.DataStore.DataContent.BlockHeaders[0].Height
	for index, blockHeader := range h.DataStore.DataContent.BlockHeaders {
		if blockHeader.Height > maxHeight {
			maxHeightIndex = index
			maxHeight = blockHeader.Height
		}
	}

	// get chainHash of block with max height
	maxheightHash := h.DataStore.DataContent.BlockHeaders[maxHeightIndex].BlockHash
	return &maxheightHash, nil
}

// func (h *MockServerHandler) GetBlock(blockHash *chainhash.Hash) (*wire.MsgBlock, error) {
// 	return in
// }

// func (h *MockServerHandler) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
// 	return in
// }

func (h *MockServerHandler) GetBlockCount() (int64, error) {
	// find the highest block height
	maxHeight := h.DataStore.DataContent.BlockHeaders[0].Height
	for _, blockHeader := range h.DataStore.DataContent.BlockHeaders {
		maxHeight = max(maxHeight, blockHeader.Height)
	}

	return maxHeight, nil
}

// func (h *MockServerHandler) GetBlockFilter(
// 	blockHash chainhash.Hash,
// 	filterType *btcjson.FilterTypeName,
// ) (*btcjson.GetBlockFilterResult, error) {
// 	return in
// }

func (h *MockServerHandler) GetBlockHash(blockHeight int64) (*chainhash.Hash, error) {
	// get chainHash of block with blockHeight
	for _, blockHeader := range h.DataStore.DataContent.BlockHeaders {
		if blockHeader.Height == blockHeight {
			return &blockHeader.BlockHash, nil
		}
	}

	return nil, &btcjson.RPCError{
		Code:    btcjson.ErrRPCOutOfRange,
		Message: "Block number out of range",
	}
}

func (h *MockServerHandler) GetBlockHeader(blockHash *chainhash.Hash) (*BlockHeader, error) {
	blockHashString := blockHash.String()

	// find the block with hash `blockHash`
	for _, blockHeader := range h.DataStore.DataContent.BlockHeaders {
		if blockHeader.BlockHash.String() == blockHashString {
			return &blockHeader, nil
		}
	}

	return nil, &btcjson.RPCError{
		Code:    btcjson.ErrRPCBlockNotFound,
		Message: "Block not found",
	}
}

// func (h *MockServerHandler) GetBlockStats(
// 	hashOrHeight interface{},
// 	stats *[]string,
// ) (*btcjson.GetBlockStatsResult, error) {
// 	return in
// }

func (h *MockServerHandler) GetTxOut(
	txHash *chainhash.Hash,
	index uint32,
	mempool bool,
) (*GetTxOutResult, error) {
	txHashString := txHash.String()
	voutIndex := index

	// find the transaction with hash `txHash`
	for _, transaction := range h.DataStore.DataContent.Transactions {
		if transaction.TxId.String() == txHashString {
			if voutIndex >= uint32(len(transaction.VOut)) {
				return nil, &btcjson.RPCError{
					Code: btcjson.ErrRPCInvalidTxVout,
					Message: "Output index number (vout) does not " +
						"exist for transaction.",
				}
			}

			txOut := &GetTxOutResult{
				BestBlock:     "", // latest block not in data/ file
				Confirmations: int64(transaction.Confirmations),
				Value:         transaction.VOut[voutIndex].Value,
				ScriptPubKey:  transaction.VOut[voutIndex].ScriptPubKey,
				Coinbase:      true, // not available in v1 "vout"
			}
			return txOut, nil
		}
	}

	// if no txn found, return error
	return nil, btcjson.NewRPCError(
		btcjson.ErrRPCNoTxInfo,
		fmt.Sprintf("No information available about transaction %v", txHash),
	)
}

// NewMockRPCServer creates a new instance of the rpcServer and starts listening
func NewMockRPCServer() *httptest.Server {
	// Create a new RPC server
	rpcServer := jsonrpc.NewServer()

	// create a handler instance and register it
	serverHandler := &MockServerHandler{}
	rpcServer.Register("MockServerHandler", serverHandler)

	// populate data from json data/ file
	serverHandler.PopulateDataStore()

	// serve the API
	testServ := httptest.NewServer(rpcServer)

	return testServ
}
