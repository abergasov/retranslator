package retranslator_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/abergasov/retranslator/pkg/testutils"
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
			resp, err := service.ProxyRequest(requestID, http.MethodGet, echoURL+"/echo/"+requestID, nil, nil, false, false)
			require.NoError(t, err)
			response := <-resp
			require.Equal(t, int32(http.StatusOK), response.StatusCode)
			require.Equal(t, requestID, string(response.Body))
		}(request)
	}
	wg.Wait()
}

func TestRetranslator_431(t *testing.T) {
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

	requestID := uuid.NewString()

	time.Sleep(1 * time.Second)
	for i := 0; i < 10000; i++ {
		resp, err := service.ProxyRequest(requestID, http.MethodGet, echoURL+"/echo/"+requestID, map[string]string{
			"X-Forwarded-For":   "A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z",
			"X-Real-IP":         "A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z",
			"X-Forwarded-Host":  "A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z",
			"X-Forwarded-Proto": "A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z",
			"X-Forwarded-Port":  "A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z",
		}, nil, false, false)
		require.NoError(t, err)
		response := <-resp
		require.Equal(t, int32(http.StatusOK), response.StatusCode)
		require.Equal(t, requestID, string(response.Body))
	}
}

func TestRetranslator_Ja3_Fingerprint(t *testing.T) {
	appPort, err := freeport.GetFreePort()
	require.NoError(t, err, "failed to get free port for app")

	service, err := testutils.NewRetranslatorServer(t, appPort)
	require.NoError(t, err)
	t.Log("retranslator server started")

	targetURL := "https://tools.scrapfly.io/api/fp/ja3?extended=1"

	address := fmt.Sprintf("127.0.0.1:%d", appPort)
	_, err = testutils.NewRetranslatorClient(t, "A", address)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	fingerPrints := make(map[string]struct{})
	for i := 0; i < 10; i++ {
		resp, err := service.ProxyRequest(uuid.NewString(), http.MethodGet, targetURL, map[string]string{}, nil, false, false)
		require.NoError(t, err)
		response := <-resp
		require.Equal(t, int32(http.StatusOK), response.StatusCode)
		type a struct {
			Digest string `json:"digest"`
		}
		var b a
		require.NoError(t, json.Unmarshal(response.Body, &b))
		require.NotEmpty(t, b.Digest)
		if _, ok := fingerPrints[b.Digest]; ok {
			t.Fatal("duplicate fingerprint")
		}
		fingerPrints[b.Digest] = struct{}{}
		println(b.Digest)
	}
}
