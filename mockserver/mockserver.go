package mockserver

import (
	"net/http/httptest"

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
	blockHash := h.DataStore.DataContent.BlockHeaders[blockHeight].BlockHash
	return &blockHash, nil
}

func (h *MockServerHandler) GetBlockHeader(blockHash *chainhash.Hash) (*BlockHeader, error) {
	blockHashString := blockHash.String()

	// find the block with hash `blockHash`
	blockIndex := 0
	for index, blockHeader := range h.DataStore.DataContent.BlockHeaders {
		if blockHeader.BlockHash.String() == blockHashString {
			blockIndex = index
		}
	}

	blockHeader := h.DataStore.DataContent.BlockHeaders[blockIndex]
	return &blockHeader, nil
}

// func (h *MockServerHandler) GetBlockStats(
// 	hashOrHeight interface{},
// 	stats *[]string,
// ) (*btcjson.GetBlockStatsResult, error) {
// 	return in
// }

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
