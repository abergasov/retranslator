package main

import (
	"log"
	"os"
	"os/signal"
	"retranslator/internal/logger"
	"retranslator/internal/service/requester/executor"
	"retranslator/internal/service/requester/orchestrator"
	"retranslator/internal/service/retranslator/client"
	"syscall"

	"go.uber.org/zap"
)

var (
	appHash    = os.Getenv("GIT_HASH")
	targetHost = "http://localhost:8080"
)

func main() {
	appLog, err := logger.NewAppLogger(appHash)
	if err != nil {
		log.Fatalf("unable to create logger: %s", err)
	}

	curler := executor.NewService(appLog.With(zap.String("service", "executor")))
	orchestra := orchestrator.NewService(appLog.With(zap.String("service", "orchestrator")), curler)

	relay := client.NewRelay(appLog.With(zap.String("service", "relay")), targetHost, orchestra)
	defer relay.Stop()

	relay.Start()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt, syscall.SIGTERM)
	<-quit
	appLog.Info("app stopping...")
}
