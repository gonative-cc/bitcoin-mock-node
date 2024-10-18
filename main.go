package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

type PingRequest struct {
	Value string `json:"value"`
}
type PingResponse struct {
	Value string `json:"value"`
}

type JSONServer struct{}

// Ping
func (t *JSONServer) Ping(r *http.Request, args *PingRequest, reply *PingResponse) error {
	pingresponse := &PingResponse{
		Value: args.Value,
	}
	*reply = *pingresponse
	return nil
}

// startRPCServer creates a new instance of the rpcServer and starts listening
func startRPCServer(config *rpcServerConfig) error {
	// Create a new RPC server
	server := rpc.NewServer()
	// Register the type of data requested as JSON
	server.RegisterCodec(json.NewCodec(), "application/json")

	// Register the service by creating a new JSON server
	server.RegisterService(new(JSONServer), "")

	// Create only / route
	router := mux.NewRouter()
	router.Handle("/", server)

	http.ListenAndServe(":"+strconv.FormatInt(config.rpcPort, 10), router)

	return nil
}

func main() {
	// set the logging level

	// read the default config file for node (node behaviour cfg)
	defaultConfig := &rpcServerConfig{
		rpcPort: 8334,
	}

	err := startRPCServer(defaultConfig)
	if err != nil {
		os.Exit(1)
	}
}
