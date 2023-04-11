package testutils

import (
	"retranslator/internal/logger"
	"retranslator/internal/service/requester/executor"
	"retranslator/internal/service/requester/orchestrator"
	"retranslator/internal/service/retranslator/client"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func NewRetranslatorClient(t *testing.T, clientID, serverAddress string) (*client.Service, error) {
	appLog, err := logger.NewAppLogger("")
	require.NoError(t, err)

	curler := executor.NewService(appLog.With(zap.String("service", "executor")))

	orchestra := orchestrator.NewService(appLog.With(zap.String("service", "orchestrator")), curler)
	t.Cleanup(func() {
		orchestra.Stop()
	})

	relay := client.NewRelay(appLog.With(zap.String("client", clientID)), serverAddress, orchestra)
	t.Cleanup(func() {
		relay.Stop()
	})

	relay.Start()

	return relay, nil
}
