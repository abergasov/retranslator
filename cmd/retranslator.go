package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/abergasov/retranslator/pkg/logger"
	"github.com/abergasov/retranslator/pkg/service/counter"
	"github.com/abergasov/retranslator/pkg/service/requester/executor"
	"github.com/abergasov/retranslator/pkg/service/requester/orchestrator"
	"github.com/abergasov/retranslator/pkg/service/retranslator/client"
	"github.com/abergasov/retranslator/pkg/storage/database"
	"go.uber.org/zap"
)

var (
	appHash = os.Getenv("GIT_HASH")
	//targetHost = "localhost:48292"
	targetHost = "165.232.118.116:48292"
	dbPath     = "storage.db"
)

func main() {
	appLog, err := logger.NewAppLogger(appHash)
	if err != nil {
		log.Fatalf("unable to create logger: %s", err)
	}
	dbConnect, err := database.InitDBConnect(dbPath)
	if err != nil {
		appLog.Fatal("unable to connect to db", err)
	}
	defer dbConnect.Close()
	requestCounter := counter.NewService(appLog.With(zap.String("service", "counter")), dbConnect)

	curler := executor.NewService(appLog.With(zap.String("service", "executor")))
	orchestra := orchestrator.NewService(appLog.With(zap.String("service", "orchestrator")), curler)

	relay := client.NewRelay(appLog.With(zap.String("service", "relay")), targetHost, orchestra, requestCounter)
	defer relay.Stop()

	relay.Start()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt, syscall.SIGTERM)
	<-quit
	appLog.Info("app stopping...")
}
