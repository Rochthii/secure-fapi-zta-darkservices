package ziti

import (
	"fmt"
	"net"
	"os"

	"github.com/openziti/sdk-golang/ziti"
)

// ZitiContext wraps the OpenZiti client context
type ZitiContext struct {
	context ziti.Context
}

// NewZitiContext loads an OpenZiti identity config file and authenticates
func NewZitiContext(identityPath string) (*ZitiContext, error) {
	if _, err := os.Stat(identityPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("identity file not found at: %s", identityPath)
	}

	cfg, err := ziti.NewConfigFromFile(identityPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ziti config: %w", err)
	}

	ctx, err := ziti.NewContext(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ziti context: %w", err)
	}

	err = ctx.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with ziti controller: %w", err)
	}

	return &ZitiContext{context: ctx}, nil
}

// Listen starts a net.Listener on the OpenZiti overlay network for the given service name
func (z *ZitiContext) Listen(serviceName string) (net.Listener, error) {
	listener, err := z.context.Listen(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on ziti service '%s': %w", serviceName, err)
	}
	return listener, nil
}

// Close releases the OpenZiti context resources
func (z *ZitiContext) Close() {
	z.context.Close()
}
