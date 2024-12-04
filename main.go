package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/filecoin-project/go-jsonrpc"

	"github.com/gonative-cc/btc-mock-node/client"
	"github.com/gonative-cc/btc-mock-node/mockserver"
)

func main() {
	mockService := mockserver.NewMockRPCServer("./data/mainnet_oldest_blocks.json")

	fmt.Printf("Mock RPC server running at: %s\n", mockService.URL)

	ctx := context.Background()
	client_handler := client.Client{}
	close_handler, err := jsonrpc.NewClient(ctx, mockService.URL, "MockServerHandler", &client_handler, nil)
	if err != nil {
		panic(err)
	}

	// Create channel to listen for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan

	close_handler()
}
