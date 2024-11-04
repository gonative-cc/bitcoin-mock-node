package client

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Client struct {
	Ping              func(int) int
	GetBestBlockHash  func() (*chainhash.Hash, error)
	GetBlockCount     func() (int64, error)
	GetBlockHash      func(blockHeight int64) (*chainhash.Hash, error)
	GetBlockHeader    func(blockHash *chainhash.Hash, verbose bool) (*btcjson.GetBlockHeaderVerboseResult, error)
	GetTxOut          func(txHash *chainhash.Hash, index uint32, mempool bool) (*btcjson.GetTxOutResult, error)
	GetRawTransaction func(txHash *chainhash.Hash, verbose bool, blockHash *chainhash.Hash) (*btcjson.TxRawResult, error)
}
