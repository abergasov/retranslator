package testutils

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type EchoServer struct {
	appAddr    string
	httpEngine *fiber.App
}

func NewEchoServer(t *testing.T, address string) (*EchoServer, error) {
	app := &EchoServer{
		appAddr:    address,
		httpEngine: fiber.New(fiber.Config{}),
	}
	app.httpEngine.Get("/echo/:requestID", func(ctx *fiber.Ctx) error {
		return ctx.SendString(ctx.Params("requestID"))
	})
	go func() {
		require.NoError(t, app.httpEngine.Listen(app.appAddr))
	}()
	t.Cleanup(func() {
		require.NoError(t, app.httpEngine.Shutdown())
	})
	return app, nil
}
