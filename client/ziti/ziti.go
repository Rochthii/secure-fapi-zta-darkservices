package ziti

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/openziti/sdk-golang/ziti"
)

type ZitiClient struct {
	context ziti.Context
}

// NewZitiClient initializes OpenZiti client context
func NewZitiClient(identityPath string) (*ZitiClient, error) {
	if _, err := os.Stat(identityPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("identity file not found at: %s", identityPath)
	}

	cfg, err := ziti.NewConfigFromFile(identityPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client ziti config: %w", err)
	}

	ctx, err := ziti.NewContext(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client ziti context: %w", err)
	}

	err = ctx.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate client with ziti controller: %w", err)
	}

	return &ZitiClient{context: ctx}, nil
}

// GetHTTPClient returns an http.Client routed entirely through the OpenZiti overlay network
func (z *ZitiClient) GetHTTPClient(serviceName string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// We ignore the network/addr parameters and dial the Ziti service directly
				return z.context.Dial(serviceName)
			},
		},
	}
}

// Close releases the Ziti context
func (z *ZitiClient) Close() {
	z.context.Close()
}
