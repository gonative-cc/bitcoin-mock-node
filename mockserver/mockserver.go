package mockserver

import (
	"net/http/httptest"

	"github.com/filecoin-project/go-jsonrpc"
)

// Have a type with some exported methods
type MockServerHandler struct{}

func (h *MockServerHandler) Ping(in int) int {
	return in
}

// NewMockRPCServer creates a new instance of the rpcServer and starts listening
func NewMockRPCServer() *httptest.Server {
	// Create a new RPC server
	rpcServer := jsonrpc.NewServer()

	// create a handler instance and register it
	serverHandler := &MockServerHandler{}
	rpcServer.Register("MockServerHandler", serverHandler)

	// serve the API
	testServ := httptest.NewServer(rpcServer)

	return testServ
}
