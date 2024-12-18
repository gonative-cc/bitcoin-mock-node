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

func (h *MockServerHandler) PopulateDataStore(dataFilePath string) {
	h.DataStore.ReadJson(dataFilePath)
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

func (h *MockServerHandler) GetBlock(
	blockHash *chainhash.Hash,
	verbosity *int,
) (*btcjson.GetBlockVerboseResult, error) {
	// NOTE: verbosity is added to be compatible with the relayer
	// the method always assumes verbosity=1

	var foundBlockHeader *btcjson.GetBlockHeaderVerboseResult = nil
	// find the block with hash `blockHash`
	if blockHeader, ok := h.DataStore.BlockHeaderBlockHashMap[blockHash.String()]; ok {
		if blockHeader.Hash == blockHash.String() {
			foundBlockHeader = &blockHeader
		}
	}

	if foundBlockHeader == nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	var foundBlockTxsIds []string
	// find transactions with blockHash
	for _, tx := range h.DataStore.DataContent.Transactions {
		if tx.BlockHash == blockHash.String() {
			foundBlockTxsIds = append(foundBlockTxsIds, tx.Hex)
		}
	}

	return &btcjson.GetBlockVerboseResult{
		Hash:          foundBlockHeader.Hash,
		Confirmations: foundBlockHeader.Confirmations,
		StrippedSize:  0, // placeholder
		Size:          0, // placeholder
		Weight:        0, // placeholder
		Height:        int64(foundBlockHeader.Height),
		Version:       foundBlockHeader.Version,
		VersionHex:    foundBlockHeader.VersionHex,
		MerkleRoot:    foundBlockHeader.MerkleRoot,
		Tx:            foundBlockTxsIds,
		Time:          foundBlockHeader.Time,
		Nonce:         uint32(foundBlockHeader.Nonce),
		Bits:          foundBlockHeader.Bits,
		Difficulty:    foundBlockHeader.Difficulty,
		PreviousHash:  foundBlockHeader.PreviousHash,
		NextHash:      foundBlockHeader.NextHash,
	}, nil
}

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
	if blockHeader, ok := h.DataStore.BlockHeaderBlockHashMap[blockHash.String()]; ok {
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

func (h *MockServerHandler) GetNetworkInfo() (*btcjson.GetNetworkInfoResult, error) {
	return &h.DataStore.DataContent.NetworkInfo, nil
}

// GetInfo returns miscellaneous info regarding the RPC server.  The returned
// info object may be void of wallet information if the remote server does
// not include wallet functionality.
// NOTE: Returns nil value as it is only usedto show that it's a btd backend to relayer
func (h *MockServerHandler) GetInfo() (*btcjson.InfoWalletResult, error) {
	return nil, nil
}

// NewMockRPCServer creates a new instance of the rpcServer and starts listening
func NewMockRPCServer(dataFilePath string) *httptest.Server {
	// Create a new RPC server
	rpcServer := jsonrpc.NewServer()

	// create a handler instance and register it
	serverHandler := &MockServerHandler{}
	rpcServer.Register("MockServerHandler", serverHandler)

	// method aliases
	rpcServer.AliasMethod("ping", "MockServerHandler.Ping")
	rpcServer.AliasMethod("getbestblockhash", "MockServerHandler.GetBestBlockHash")
	rpcServer.AliasMethod("getblock", "MockServerHandler.GetBlock")
	rpcServer.AliasMethod("getblockcount", "MockServerHandler.GetBlockCount")
	rpcServer.AliasMethod("getblockhash", "MockServerHandler.GetBlockHash")
	rpcServer.AliasMethod("getblockheader", "MockServerHandler.GetBlockHeader")
	rpcServer.AliasMethod("gettxout", "MockServerHandler.GetTxOut")
	rpcServer.AliasMethod("getrawtransaction", "MockServerHandler.GetRawTransaction")
	rpcServer.AliasMethod("getnetworkinfo", "MockServerHandler.GetNetworkInfo")
	rpcServer.AliasMethod("getinfo", "MockServerHandler.GetInfo")

	// populate data from json data/ file
	serverHandler.PopulateDataStore(dataFilePath)

	// serve the API
	testServ := httptest.NewServer(rpcServer)

	return testServ
}
