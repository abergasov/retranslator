package counter

import (
	"fmt"
	"sync"
	"time"

	"github.com/abergasov/retranslator/pkg/logger"
	"github.com/abergasov/retranslator/pkg/storage/database"
)

const (
	tableName   = "counter_stat"
	maxRequests = 200_000
)

// Service tracks the number of requests from specific IP addresses.
type Service struct {
	log         logger.AppLogger
	currentIPMU sync.RWMutex
	currentIP   string

	// counters requests are here
	counter     map[string]map[string]uint64
	counterMU   sync.Mutex
	currentDate string

	conn database.DBConnector
}

func NewService(log logger.AppLogger, db database.DBConnector) *Service {
	srv := &Service{
		log:         log,
		conn:        db,
		currentDate: time.Now().Format(time.DateOnly),
		counter:     make(map[string]map[string]uint64),
	}
	go srv.observeIP()
	go srv.backupState()
	if err := srv.migrate(); err != nil {
		srv.log.Error("unable to migrate", err)
	}
	if err := srv.loadState(); err != nil {
		srv.log.Error("unable to load state", err)
	}
	return srv
}

func (s *Service) CanRequest() error {
	s.currentIPMU.RLock()
	ip := s.currentIP
	s.currentIPMU.RUnlock()
	s.counterMU.Lock()
	defer s.counterMU.Unlock()
	if _, ok := s.counter[s.currentDate]; !ok {
		s.counter[s.currentDate] = make(map[string]uint64)
	}
	if _, ok := s.counter[s.currentDate][ip]; !ok {
		s.counter[s.currentDate][ip] = 0
	}
	s.counter[s.currentDate][ip]++
	if s.counter[s.currentDate][ip] > maxRequests {
		return fmt.Errorf("max requests exceeded")
	}
	return nil
}

func (s *Service) Stop() {
	if err := s.saveState(); err != nil {
		s.log.Error("failed save state on exist", err)
	}
}
