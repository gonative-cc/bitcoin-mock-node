package mockserver

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/gonative-cc/btc-mock-node/client"
	"github.com/stretchr/testify/assert"
)

// setup initializes the test instance and sets up common resources.
func setup(t *testing.T) (client.Client, jsonrpc.ClientCloser) {
	mockService := NewMockRPCServer()

	t.Logf("mock json-rpc server listening on: %s", mockService.URL)

	ctx := context.Background()
	client_handler := client.Client{}

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

	t.Run("GetBlockHashError", func(t *testing.T) {
		_, err := client_handler.GetBlockHash(15)
		assert.Error(t, err)
	})

	t.Run("GetBlockHeader", func(t *testing.T) {
		blockHash, err := chainhash.NewHashFromStr("0000000071966c2b1d065fd446b1e485b2c9d9594acd2007ccbd5441cfc89444")
		assert.NoError(t, err)

		blockHeader, err := client_handler.GetBlockHeader(blockHash, true)
		assert.NoError(t, err)

		// https://learnmeabitcoin.com/explorer/block/0000000071966c2b1d065fd446b1e485b2c9d9594acd2007ccbd5441cfc89444
		actualBlockHeader := &btcjson.GetBlockHeaderVerboseResult{
			Hash:          "0000000071966c2b1d065fd446b1e485b2c9d9594acd2007ccbd5441cfc89444",
			Confirmations: 867297,
			Height:        7,
			Version:       1,
			VersionHex:    "00000001",
			MerkleRoot:    "8aa673bc752f2851fd645d6a0a92917e967083007d9c1684f9423b100540673f",
			Time:          1231472369,
			// MedianTime:    1231470988,
			Nonce:      2258412857,
			Bits:       "1d00ffff",
			Difficulty: 1,
			// ChainWork:     *actualChainWork,
			// NumberTx:      1,
			PreviousHash: "000000003031a0e73735690c5a1ff2a4be82553b2a12b776fbd3a215dc8f778d",
			NextHash:     "00000000408c48f847aa786c2268fc3e6ec2af68e8468a34a28c61b7f1de0dc6",
		}

		assert.Equal(t, actualBlockHeader, blockHeader)
	})

	t.Run("GetBlockHeaderError", func(t *testing.T) {
		blockHash, err := chainhash.NewHashFromStr("0000000071966c2b1d065fd446b1e485b2c9d9594acd2007ccbd5441cfc89222")
		assert.NoError(t, err)

		_, err = client_handler.GetBlockHeader(blockHash, true)
		assert.Error(t, err)
	})

	t.Run("GetTxOut", func(t *testing.T) {
		txnHash, err := chainhash.NewHashFromStr("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098")
		assert.NoError(t, err)

		txOut, err := client_handler.GetTxOut(txnHash, 0, false)
		assert.NoError(t, err)

		// https://learnmeabitcoin.com/explorer/tx/0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098
		actualTxOut := &btcjson.GetTxOutResult{
			BestBlock:     "",
			Confirmations: 867743,
			Value:         50,
			ScriptPubKey: btcjson.ScriptPubKeyResult{
				Asm:  "0496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858ee OP_CHECKSIG",
				Hex:  "410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac",
				Type: "pubkey",
			},
			Coinbase: true,
		}
		assert.NoError(t, err)

		assert.Equal(t, actualTxOut, txOut)
	})

	t.Run("GetTxOutIncorrectHash", func(t *testing.T) {
		txnHash, err := chainhash.NewHashFromStr("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512099")
		assert.NoError(t, err)

		_, err = client_handler.GetTxOut(txnHash, 0, false)
		assert.Error(t, err)
	})

	t.Run("GetTxOutOutOfIndex", func(t *testing.T) {
		txnHash, err := chainhash.NewHashFromStr("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098")
		assert.NoError(t, err)

		_, err = client_handler.GetTxOut(txnHash, 1, false)
		assert.Error(t, err)
	})

	t.Run("GetRawTransaction", func(t *testing.T) {
		txnHash, err := chainhash.NewHashFromStr("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098")
		assert.NoError(t, err)

		rawTx, err := client_handler.GetRawTransaction(txnHash, true, nil)
		assert.NoError(t, err)

		// https://learnmeabitcoin.com/explorer/tx/0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098
		actualRawTx := &btcjson.TxRawResult{
			Txid:     "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
			Hash:     "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
			Version:  1,
			Size:     134,
			Vsize:    134,
			Weight:   536,
			LockTime: 0,
			Vin: []btcjson.Vin{
				{
					Coinbase: "04ffff001d0104",
					Sequence: 4294967295,
				},
			},
			Vout: []btcjson.Vout{
				{
					Value: 50,
					N:     0,
					ScriptPubKey: btcjson.ScriptPubKeyResult{
						Asm:  "0496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858ee OP_CHECKSIG",
						Hex:  "410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac",
						Type: "pubkey",
					},
				},
			},
			Hex:           "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0704ffff001d0104ffffffff0100f2052a0100000043410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac00000000",
			BlockHash:     "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048",
			Confirmations: 867743,
			Time:          1231469665,
			Blocktime:     1231469665,
		}
		assert.NoError(t, err)

		assert.Equal(t, actualRawTx, rawTx)
	})

	t.Run("GetRawTransactionIncorrectHash", func(t *testing.T) {
		txnHash, err := chainhash.NewHashFromStr("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512099")
		assert.NoError(t, err)

		_, err = client_handler.GetRawTransaction(txnHash, true, nil)
		assert.Error(t, err)
	})
}
