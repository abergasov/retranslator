package testutils

import (
	"fmt"
	"net"
	"retranslator/internal/service/retranslator/server"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func NewRetranslatorServer(t *testing.T, appPort int) (*server.Service, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", appPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	grpcServer := grpc.NewServer()
	srv := server.NewService()
	srv.Start(grpcServer)
	t.Cleanup(func() {
		srv.Stop()
		require.NoError(t, listener.Close())
	})
	go func() {
		require.NoError(t, grpcServer.Serve(listener))
	}()
	return srv, nil
}
