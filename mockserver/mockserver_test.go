package mockserver

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/stretchr/testify/assert"
)

type Client struct {
	Ping             func(int) int
	GetBestBlockHash func() (*chainhash.Hash, error)
	GetBlockCount    func() (int64, error)
	GetBlockHash     func(blockHeight int64) (*chainhash.Hash, error)
	GetBlockHeader   func(blockHash *chainhash.Hash) (*BlockHeader, error)
}

// setup initializes the test instance and sets up common resources.
func setup(t *testing.T) (Client, jsonrpc.ClientCloser) {
	mockService := NewMockRPCServer()

	t.Logf("mock json-rpc server listening on: %s", mockService.URL)

	ctx := context.Background()
	client_handler := Client{}

	close_handler, err := jsonrpc.NewClient(ctx, mockService.URL, "MockServerHandler", &client_handler, nil)
	assert.NoError(t, err)

	return client_handler, close_handler
}

// teardown closes the client
func teardown(close_handler jsonrpc.ClientCloser) {
	close_handler()
}

func TestMockRPCServer(t *testing.T) {
	client_handler, close_handler := setup(t)
	defer teardown(close_handler)

	t.Run("Ping", func(t *testing.T) {
		pingValue := client_handler.Ping(10)
		assert.Equal(t, 10, pingValue)
	})

	t.Run("GetBestBlockHash", func(t *testing.T) {
		bestBlockHash, err := client_handler.GetBestBlockHash()
		assert.NoError(t, err)

		actualBlockHash, err := chainhash.NewHashFromStr("000000002c05cc2e78923c34df87fd108b22221ac6076c18f3ade378a4d915e9")
		assert.NoError(t, err)

		assert.Equal(t, actualBlockHash, bestBlockHash)
	})

	t.Run("GetBlockCount", func(t *testing.T) {
		blockCount, err := client_handler.GetBlockCount()
		assert.NoError(t, err)

		assert.Equal(t, int64(10), blockCount)
	})

	t.Run("GetBlockHash", func(t *testing.T) {
		blockHash, err := client_handler.GetBlockHash(5)
		assert.NoError(t, err)

		actualBlockHash, err := chainhash.NewHashFromStr("000000009b7262315dbf071787ad3656097b892abffd1f95a1a022f896f533fc")
		assert.NoError(t, err)

		assert.Equal(t, actualBlockHash, blockHash)
	})

	t.Run("GetBlockHeader", func(t *testing.T) {
		blockHash, err := chainhash.NewHashFromStr("0000000071966c2b1d065fd446b1e485b2c9d9594acd2007ccbd5441cfc89444")
		assert.NoError(t, err)

		blockHeader, err := client_handler.GetBlockHeader(blockHash)
		assert.NoError(t, err)

		actualBlockHash, err := chainhash.NewHashFromStr("0000000071966c2b1d065fd446b1e485b2c9d9594acd2007ccbd5441cfc89444")
		assert.NoError(t, err)
		actualMerkleRoot, err := chainhash.NewHashFromStr("8aa673bc752f2851fd645d6a0a92917e967083007d9c1684f9423b100540673f")
		assert.NoError(t, err)
		actualChainWork, err := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000800080008")
		assert.NoError(t, err)
		actualPrevBlock, err := chainhash.NewHashFromStr("000000003031a0e73735690c5a1ff2a4be82553b2a12b776fbd3a215dc8f778d")
		assert.NoError(t, err)
		actualNextBlock, err := chainhash.NewHashFromStr("00000000408c48f847aa786c2268fc3e6ec2af68e8468a34a28c61b7f1de0dc6")
		assert.NoError(t, err)
		actualBlockHeader := &BlockHeader{
			BlockHash:     *actualBlockHash,
			Confirmations: 867297,
			Height:        7,
			Version:       1,
			VersionHex:    "00000001",
			MerkleRoot:    *actualMerkleRoot,
			Time:          1231472369,
			MedianTime:    1231470988,
			Nonce:         2258412857,
			Bits:          "1d00ffff",
			Difficulty:    1,
			ChainWork:     *actualChainWork,
			NumberTx:      1,
			PrevBlock:     *actualPrevBlock,
			NextBlock:     *actualNextBlock,
		}

		assert.Equal(t, actualBlockHeader, blockHeader)
	})
}
