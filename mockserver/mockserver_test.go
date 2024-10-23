package mockserver

import (
	"context"
	"testing"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/stretchr/testify/assert"
)

type Client struct {
	Ping func(int) int
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
}
