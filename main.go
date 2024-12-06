package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/rs/zerolog/log"

	"github.com/gonative-cc/btc-mock-node/client"
	"github.com/gonative-cc/btc-mock-node/mockserver"
)

func main() {
	// input path of json data file as cli argument
	// example: ./data/mainnet_oldest_blocks.json
	if len(os.Args) < 2 {
		log.Error().Msg("Missing transaction file path")
		return
	}
	txFilePath := os.Args[1]

	mockService := mockserver.NewMockRPCServer(txFilePath)

	log.Info().Msgf("Mock RPC server running at: %s", mockService.URL)

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
