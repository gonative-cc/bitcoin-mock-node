package main

import (
	"crypto/sha256"
	"encoding/base64"
	"net"
	"sync"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/blockchain/indexers"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/mining"
	"github.com/btcsuite/btcd/mining/cpuminer"
	"github.com/btcsuite/btcd/wire"
	"google.golang.org/grpc/peer"
)

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate code set.  It also logs the error to the
// RPC server subsystem since internal errors really should not occur.  The
// context parameter is only used in the log message and may be empty if it's
// not needed.
func internalRPCError(errStr, context string) *btcjson.RPCError {
	// TODO: uncomment and add logger
	// logStr := errStr
	// if context != "" {
	// 	logStr = context + ": " + errStr
	// }
	// rpcsLog.Error(logStr)
	return btcjson.NewRPCError(btcjson.ErrRPCInternal.Code, errStr)
}

type commandHandler func(*rpcServer, interface{}, <-chan struct{}) (interface{}, error)

// rpcHandlers maps RPC command strings to appropriate handler functions.
// This is set by init because help references rpcHandlers and thus causes
// a dependency loop.
var rpcHandlers = map[string]commandHandler{
	"ping": handlePing,
	// "getheaders": handleGetHeaders,
}

// standardCmdResult checks that a parsed command is a standard Bitcoin JSON-RPC
// command and runs the appropriate handler to reply to the command.  Any
// commands which are not recognized or not implemented will return an error
// suitable for use in replies.
func (s *rpcServer) standardCmdResult(
	cmd *parsedRPCCmd,
	closeChan <-chan struct{},
) (interface{}, error) {
	handler, ok := rpcHandlers[cmd.method]
	if ok {
		return handler(s, cmd.cmd, closeChan)
	}
	return nil, btcjson.ErrRPCMethodNotFound
}

// createMarshalledReply returns a new marshalled JSON-RPC response given the
// passed parameters.  It will automatically convert errors that are not of
// the type *btcjson.RPCError to the appropriate type as needed.
func createMarshalledReply(
	rpcVersion btcjson.RPCVersion,
	id interface{},
	result interface{},
	replyErr error,
) ([]byte, error) {
	var jsonErr *btcjson.RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*btcjson.RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = internalRPCError(replyErr.Error(), "")
		}
	}

	return btcjson.MarshalResponse(rpcVersion, id, result, jsonErr)
}

// parseCmd parses a JSON-RPC request object into known concrete command.  The
// err field of the returned parsedRPCCmd struct will contain an RPC error that
// is suitable for use in replies if the command is invalid in some way such as
// an unregistered command or invalid parameters.
func parseCmd(request *btcjson.Request) *parsedRPCCmd {
	parsedCmd := parsedRPCCmd{
		jsonrpc: request.Jsonrpc,
		id:      request.ID,
		method:  request.Method,
	}

	cmd, err := btcjson.UnmarshalCmd(request)
	if err != nil {
		// When the error is because the method is not registered,
		// produce a method not found RPC error.
		if jerr, ok := err.(btcjson.Error); ok &&
			jerr.ErrorCode == btcjson.ErrUnregisteredMethod {

			parsedCmd.err = btcjson.ErrRPCMethodNotFound
			return &parsedCmd
		}

		// Otherwise, some type of invalid parameters is the
		// cause, so produce the equivalent RPC error.
		parsedCmd.err = btcjson.NewRPCError(
			btcjson.ErrRPCInvalidParams.Code, err.Error())
		return &parsedCmd
	}

	parsedCmd.cmd = cmd
	return &parsedCmd
}

// processRequest determines the incoming request type (single or batched),
// parses it and returns a marshalled response.
func (s *rpcServer) processRequest(
	request *btcjson.Request,
	isAdmin bool,
	closeChan <-chan struct{},
) []byte {
	var result interface{}
	var err error
	var jsonErr *btcjson.RPCError

	if jsonErr == nil {
		if request.Method == "" || request.Params == nil {
			jsonErr = &btcjson.RPCError{
				Code:    btcjson.ErrRPCInvalidRequest.Code,
				Message: "Invalid request: malformed",
			}
			msg, err := createMarshalledReply(request.Jsonrpc, request.ID, result, jsonErr)
			if err != nil {
				// TODO: uncomment and add logger
				// rpcsLog.Errorf("Failed to marshal reply: %v", err)
				return nil
			}
			return msg
		}

		// Valid requests with no ID (notifications) must not have a response
		// per the JSON-RPC spec.
		if request.ID == nil {
			return nil
		}

		// Attempt to parse the JSON-RPC request into a known
		// concrete command.
		parsedCmd := parseCmd(request)
		if parsedCmd.err != nil {
			jsonErr = parsedCmd.err
		} else {
			result, err = s.standardCmdResult(parsedCmd,
				closeChan)
			if err != nil {
				if rpcErr, ok := err.(*btcjson.RPCError); ok {
					jsonErr = rpcErr
				} else {
					jsonErr = &btcjson.RPCError{
						Code:    btcjson.ErrRPCInvalidRequest.Code,
						Message: "Invalid request: malformed",
					}
				}
			}
		}
	}

	// Marshal the response.
	msg, err := createMarshalledReply(request.Jsonrpc, request.ID, result, jsonErr)
	if err != nil {
		// TODO: uncomment and add logger
		// rpcsLog.Errorf("Failed to marshal reply: %v", err)
		return nil
	}
	return msg
}

// // handleGetHeaders implements the getheaders command.
// //
// // NOTE: This is a btcsuite extension originally ported from
// // github.com/decred/dcrd.
// func handleGetHeaders(
// 	s *rpcServer, cmd interface{},
// 	closeChan <-chan struct{},
// ) (interface{}, error) {
// 	c := cmd.(*btcjson.GetHeadersCmd)

// 	// Fetch the requested headers from chain while respecting the provided
// 	// block locators and stop hash.
// 	blockLocators := make([]*chainhash.Hash, len(c.BlockLocators))
// 	for i := range c.BlockLocators {
// 		blockLocator, err := chainhash.NewHashFromStr(c.BlockLocators[i])
// 		if err != nil {
// 			return nil, rpcDecodeHexError(c.BlockLocators[i])
// 		}
// 		blockLocators[i] = blockLocator
// 	}
// 	var hashStop chainhash.Hash
// 	if c.HashStop != "" {
// 		err := chainhash.Decode(&hashStop, c.HashStop)
// 		if err != nil {
// 			return nil, rpcDecodeHexError(c.HashStop)
// 		}
// 	}
// 	headers := s.cfg.SyncMgr.LocateHeaders(blockLocators, &hashStop)

// 	// Return the serialized block headers as hex-encoded strings.
// 	hexBlockHeaders := make([]string, len(headers))
// 	var buf bytes.Buffer
// 	for i, h := range headers {
// 		err := h.Serialize(&buf)
// 		if err != nil {
// 			return nil, internalRPCError(err.Error(),
// 				"Failed to serialize block header")
// 		}
// 		hexBlockHeaders[i] = hex.EncodeToString(buf.Bytes())
// 		buf.Reset()
// 	}
// 	return hexBlockHeaders, nil
// }

// handlePing implements the ping command.
func handlePing(
	s *rpcServer,
	cmd interface{},
	closeChan <-chan struct{},
) (interface{}, error) {
	// Ask server to ping \o_
	nonce, err := wire.RandomUint64()
	if err != nil {
		return nil, internalRPCError("Not sending ping - failed to "+
			"generate nonce: "+err.Error(), "")
	}
	s.cfg.ConnMgr.BroadcastMessage(wire.NewMsgPing(nonce))

	return nil, nil
}

// rpcserverPeer represents a peer for use with the RPC server.
//
// The interface contract requires that all of these methods are safe for
// concurrent access.
type rpcserverPeer interface {
	// ToPeer returns the underlying peer instance.
	ToPeer() *peer.Peer

	// IsTxRelayDisabled returns whether or not the peer has disabled
	// transaction relay.
	IsTxRelayDisabled() bool

	// BanScore returns the current integer value that represents how close
	// the peer is to being banned.
	BanScore() uint32

	// FeeFilter returns the requested current minimum fee rate for which
	// transactions should be announced.
	FeeFilter() int64
}

// rpcserverConnManager represents a connection manager for use with the RPC
// server.
//
// The interface contract requires that all of these methods are safe for
// concurrent access.
type rpcserverConnManager interface {
	// Connect adds the provided address as a new outbound peer.  The
	// permanent flag indicates whether or not to make the peer persistent
	// and reconnect if the connection is lost.  Attempting to connect to an
	// already existing peer will return an error.
	Connect(addr string, permanent bool) error

	// RemoveByID removes the peer associated with the provided id from the
	// list of persistent peers.  Attempting to remove an id that does not
	// exist will return an error.
	RemoveByID(id int32) error

	// RemoveByAddr removes the peer associated with the provided address
	// from the list of persistent peers.  Attempting to remove an address
	// that does not exist will return an error.
	RemoveByAddr(addr string) error

	// DisconnectByID disconnects the peer associated with the provided id.
	// This applies to both inbound and outbound peers.  Attempting to
	// remove an id that does not exist will return an error.
	DisconnectByID(id int32) error

	// DisconnectByAddr disconnects the peer associated with the provided
	// address.  This applies to both inbound and outbound peers.
	// Attempting to remove an address that does not exist will return an
	// error.
	DisconnectByAddr(addr string) error

	// ConnectedCount returns the number of currently connected peers.
	ConnectedCount() int32

	// NetTotals returns the sum of all bytes received and sent across the
	// network for all peers.
	NetTotals() (uint64, uint64)

	// ConnectedPeers returns an array consisting of all connected peers.
	ConnectedPeers() []rpcserverPeer

	// PersistentPeers returns an array consisting of all the persistent
	// peers.
	PersistentPeers() []rpcserverPeer

	// BroadcastMessage sends the provided message to all currently
	// connected peers.
	BroadcastMessage(msg wire.Message)

	// AddRebroadcastInventory adds the provided inventory to the list of
	// inventories to be rebroadcast at random intervals until they show up
	// in a block.
	AddRebroadcastInventory(iv *wire.InvVect, data interface{})

	// RelayTransactions generates and relays inventory vectors for all of
	// the passed transactions to all connected peers.
	RelayTransactions(txns []*mempool.TxDesc)

	// NodeAddresses returns an array consisting node addresses which can
	// potentially be used to find new nodes in the network.
	NodeAddresses() []*wire.NetAddressV2
}

// rpcserverConfig is a descriptor containing the RPC server configuration.
type rpcserverConfig struct {
	// Listeners defines a slice of listeners for which the RPC server will
	// take ownership of and accept connections.  Since the RPC server takes
	// ownership of these listeners, they will be closed when the RPC server
	// is stopped.
	Listeners []net.Listener

	// StartupTime is the unix timestamp for when the server that is hosting
	// the RPC server started.
	StartupTime int64

	// ConnMgr defines the connection manager for the RPC server to use.  It
	// provides the RPC server with a means to do things such as add,
	// remove, connect, disconnect, and query peers as well as other
	// connection-related data and tasks.
	ConnMgr rpcserverConnManager

	// // SyncMgr defines the sync manager for the RPC server to use.
	// SyncMgr rpcserverSyncManager

	// These fields allow the RPC server to interface with the local block
	// chain data and state.
	TimeSource  blockchain.MedianTimeSource
	Chain       *blockchain.BlockChain
	ChainParams *chaincfg.Params
	DB          database.DB

	// TxMemPool defines the transaction memory pool to interact with.
	TxMemPool mempool.TxMempool

	// These fields allow the RPC server to interface with mining.
	//
	// Generator produces block templates and the CPUMiner solves them using
	// the CPU.  CPU mining is typically only useful for test purposes when
	// doing regression or simulation testing.
	Generator *mining.BlkTmplGenerator
	CPUMiner  *cpuminer.CPUMiner

	// These fields define any optional indexes the RPC server can make use
	// of to provide additional data when queried.
	TxIndex   *indexers.TxIndex
	AddrIndex *indexers.AddrIndex
	CfIndex   *indexers.CfIndex

	// The fee estimator keeps track of how long transactions are left in
	// the mempool before they are mined into blocks.
	FeeEstimator *mempool.FeeEstimator
}

// rpcServer provides a concurrent safe RPC server to a chain server.
type rpcServer struct {
	started      int32
	shutdown     int32
	cfg          rpcserverConfig
	authsha      [sha256.Size]byte
	limitauthsha [sha256.Size]byte
	// ntfnMgr                *wsNotificationManager
	numClients  int32
	statusLines map[int]string
	statusLock  sync.RWMutex
	wg          sync.WaitGroup
	// gbtWorkState           *gbtWorkState
	// helpCacher             *helpCacher
	requestProcessShutdown chan struct{}
	quit                   chan int
}

// parsedRPCCmd represents a JSON-RPC request object that has been parsed into
// a known concrete command along with any error that might have happened while
// parsing it.
type parsedRPCCmd struct {
	jsonrpc btcjson.RPCVersion
	id      interface{}
	method  string
	cmd     interface{}
	err     *btcjson.RPCError
}

// newRPCServer returns a new instance of the rpcServer struct.
func newRPCServer(config *rpcserverConfig) (*rpcServer, error) {
	rpc := rpcServer{
		cfg:         *config,
		statusLines: make(map[int]string),
		// gbtWorkState:           newGbtWorkState(config.TimeSource),
		// helpCacher:             newHelpCacher(),
		requestProcessShutdown: make(chan struct{}),
		quit:                   make(chan int),
	}
	// if cfg.RPCUser != "" && cfg.RPCPass != "" {
	// login := cfg.RPCUser + ":" + cfg.RPCPass
	login := "RPCUser" + ":" + "RPCPass"
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
	rpc.authsha = sha256.Sum256([]byte(auth))
	// }
	// if cfg.RPCLimitUser != "" && cfg.RPCLimitPass != "" {
	// login := cfg.RPCLimitUser + ":" + cfg.RPCLimitPass
	loginLimit := "RPCLimitUser" + ":" + "RPCLimitPass"
	authLimit := "Basic " + base64.StdEncoding.EncodeToString([]byte(loginLimit))
	rpc.limitauthsha = sha256.Sum256([]byte(authLimit))
	// }
	// rpc.ntfnMgr = newWsNotificationManager(&rpc)
	// rpc.cfg.Chain.Subscribe(rpc.handleBlockchainNotification)

	return &rpc, nil
}
