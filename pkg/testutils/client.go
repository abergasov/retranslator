package testutils

import (
	"testing"

	"github.com/abergasov/retranslator/pkg/logger"
	"github.com/abergasov/retranslator/pkg/service/requester/executor"
	"github.com/abergasov/retranslator/pkg/service/requester/orchestrator"
	"github.com/abergasov/retranslator/pkg/service/retranslator/client"
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
