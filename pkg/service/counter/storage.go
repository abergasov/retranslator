package counter

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type event struct {
	Date      string    `db:"retranslation_date"`
	IP        string    `db:"used_ip"`
	Count     uint64    `db:"total_counts"`
	LastUsage time.Time `db:"-"`
	LastI     string    `db:"last_update"`
}

func (e *event) parse() {
	e.LastUsage, _ = time.Parse(time.DateTime, e.LastI)
}

func (s *Service) migrate() error {
	q := []string{fmt.Sprintf(`create table if not exists %s
		(
			retranslation_date text,
			used_ip            text,
			total_counts       integer,
			last_update		   text,
			constraint counter_stat_pk
				primary key (retranslation_date, used_ip)
		);`, tableName),
		fmt.Sprintf(`delete from %s where retranslation_date != '%s'`, tableName, s.currentDate),
	}
	for _, query := range q {
		if _, err := s.conn.Client().Exec(query); err != nil {
			return fmt.Errorf("unable to migrate: %w", err)
		}
	}
	return nil
}

func (s *Service) loadState() error {
	rows, err := s.conn.Client().Queryx(fmt.Sprintf(`SELECT * FROM %s`, tableName))
	if err != nil {
		return fmt.Errorf("unable to run query: %w", err)
	}
	defer rows.Close()
	data := make(map[string]map[string]uint64)
	for rows.Next() {
		var evt event
		if err = rows.StructScan(&evt); err != nil {
			return fmt.Errorf("unable to scan row: %w", err)
		}
		evt.parse()
		if _, ok := data[evt.Date]; !ok {
			data[evt.Date] = make(map[string]uint64)
		}
		data[evt.Date][evt.IP] = evt.Count
		s.log.Info("load processed requests", zap.String("date", evt.Date), zap.String("ip", evt.IP), zap.Uint64("count", evt.Count))
	}
	s.counterMU.Lock()
	s.counter = data
	s.counterMU.Unlock()
	return nil
}

func (s *Service) backupState() {
	for range time.NewTicker(5 * time.Minute).C {
		if err := s.saveState(); err != nil {
			s.log.Error("error update counter state", err)
		}
	}
}

func (s *Service) saveState() error {
	s.counterMU.Lock()
	state := s.counter
	s.counterMU.Unlock()
	for date, ips := range state {
		for ip, count := range ips {
			s.log.Info("saving state", zap.Uint64("count", count), zap.String("ip", ip))
			sql := fmt.Sprintf(`INSERT INTO %s (retranslation_date, used_ip, total_counts, last_update) VALUES (?, ?, ?, ?) ON CONFLICT (retranslation_date, used_ip) DO UPDATE SET total_counts = ?, last_update = ?;`, tableName)
			if _, err := s.conn.Client().Exec(sql, date, ip, count, time.Now().Format(time.DateTime), count, time.Now().Format(time.DateTime)); err != nil {
				return fmt.Errorf("unable to save state: %w", err)
			}
		}
	}
	s.log.Info("state saved")
	return nil
}
