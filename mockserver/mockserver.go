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
	maxheightHash := h.DataStore.DataContent.BlockHeaders[maxHeightIndex].Hash

	bestBlockHash, err := chainhash.NewHashFromStr(maxheightHash)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDecodeHexString,
			Message: "Unable to parse block hash stored",
		}
	}
	return bestBlockHash, nil
}

// func (h *MockServerHandler) GetBlock(blockHash *chainhash.Hash) (*wire.MsgBlock, error) {
// 	return in
// }

func (h *MockServerHandler) GetBlockCount() (int32, error) {
	// find the highest block height
	maxHeight := h.DataStore.DataContent.BlockHeaders[0].Height
	for _, blockHeader := range h.DataStore.DataContent.BlockHeaders {
		maxHeight = max(maxHeight, blockHeader.Height)
	}

	return maxHeight, nil
}

func (h *MockServerHandler) GetBlockHash(blockHeight int32) (*chainhash.Hash, error) {
	// get chainHash of block with blockHeight
	if blockHeader, ok := h.DataStore.BlockHeaderMap[blockHeight]; ok {
		blockHash, err := chainhash.NewHashFromStr(blockHeader.Hash)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCDecodeHexString,
				Message: "Unable to parse block hash stored",
			}
		}
		return blockHash, nil
	}

	return nil, &btcjson.RPCError{
		Code:    btcjson.ErrRPCOutOfRange,
		Message: "Block number out of range",
	}
}

func (h *MockServerHandler) GetBlockHeader(
	blockHash *chainhash.Hash,
	verbose bool,
) (*btcjson.GetBlockHeaderVerboseResult, error) {
	// find the block with hash `blockHash`
	for _, blockHeader := range h.DataStore.DataContent.BlockHeaders {
		if blockHeader.Hash == blockHash.String() {
			return &blockHeader, nil
		}
	}

	return nil, &btcjson.RPCError{
		Code:    btcjson.ErrRPCBlockNotFound,
		Message: "Block not found",
	}
}

func (h *MockServerHandler) GetTxOut(
	txHash *chainhash.Hash,
	index uint32,
	mempool bool,
) (*btcjson.GetTxOutResult, error) {
	voutIndex := index

	// find the transaction with hash `txHash`
	if transaction, ok := h.DataStore.TransactionMap[txHash.String()]; ok {
		if voutIndex >= uint32(len(transaction.Vout)) {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInvalidTxVout,
				Message: "Output index number (vout) does not " +
					"exist for transaction.",
			}
		}

		txOut := &btcjson.GetTxOutResult{
			BestBlock:     "", // latest block not in data/ file
			Confirmations: int64(transaction.Confirmations),
			Value:         transaction.Vout[voutIndex].Value,
			ScriptPubKey:  transaction.Vout[voutIndex].ScriptPubKey,
			Coinbase:      true, // not available in v1 "vout"
		}
		return txOut, nil
	}

	// if no txn found, return error
	return nil, btcjson.NewRPCError(
		btcjson.ErrRPCNoTxInfo,
		fmt.Sprintf("No information available about transaction %v", txHash),
	)
}

func (h *MockServerHandler) GetRawTransaction(
	txHash *chainhash.Hash,
	verbose bool,
	blockHash *chainhash.Hash,
) (*btcjson.TxRawResult, error) {
	// find the transaction with hash `txHash`
	if transaction, ok := h.DataStore.TransactionMap[txHash.String()]; ok {
		return &transaction, nil
	}

	return nil, &btcjson.RPCError{
		Code:    btcjson.ErrRPCRawTxString,
		Message: "Transaction not found",
	}
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
