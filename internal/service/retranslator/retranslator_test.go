package retranslator_test

import (
	"fmt"
	"net/http"
	"retranslator/internal/testutils"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

const (
	totalRequests = 100
)

func TestRetranslator(t *testing.T) {
	appPort, err := freeport.GetFreePort()
	require.NoError(t, err, "failed to get free port for app")

	service, err := testutils.NewRetranslatorServer(t, appPort)
	require.NoError(t, err)
	t.Log("retranslator server started")

	appPortEcho, err := freeport.GetFreePort()
	require.NoError(t, err, "failed to get free port for app")

	echoURL := fmt.Sprintf("http://127.0.0.1:%d", appPortEcho)
	_, err = testutils.NewEchoServer(t, fmt.Sprintf(":%d", appPortEcho))
	require.NoError(t, err)

	address := fmt.Sprintf("127.0.0.1:%d", appPort)
	_, err = testutils.NewRetranslatorClient(t, "A", address)
	require.NoError(t, err)
	_, err = testutils.NewRetranslatorClient(t, "B", address)
	require.NoError(t, err)

	requests := make([]string, 0, totalRequests)
	for i := 0; i < totalRequests; i++ {
		requests = append(requests, uuid.NewString())
	}
	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup
	for _, request := range requests {
		wg.Add(1)
		go func(requestID string) {
			defer wg.Done()
			resp, err := service.ProxyRequest(requestID, http.MethodGet, echoURL+"/echo/"+requestID, nil, false, false)
			require.NoError(t, err)
			response := <-resp
			require.Equal(t, int32(http.StatusOK), response.StatusCode)
			require.Equal(t, requestID, string(response.Body))
		}(request)
	}
	wg.Wait()
}
