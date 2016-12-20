package core

import (
	"database/sql"
	"log"
	"time"

	"github.com/lib/pq"
)

// Stats is base statistics gatherer
type Stats struct {
	Start      pq.NullTime
	Stop       pq.NullTime
	Size       int64
	ExitCode   int
	Stacktrace string
}

// NewStats creates Stats object
func NewStats() *Stats {
	return &Stats{Size: 0}
}

// Duration calculates run duration
func (s *Stats) Duration() time.Duration {
	return s.Stop.Time.Sub(s.Start.Time) / time.Millisecond
}

// Send writes statistics data to database
func (s *Stats) Send(dbURL string, serverID int) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error connecting to database: %s\n", err)
		return err
	}
	defer db.Close()
	_, err = db.Exec(`INSERT INTO server_run_statistics (start, stop, size, exit_code, stacktrace, server_id)
	 VALUES ($1, $2, $3, $4, $5, $6)`, s.Start, s.Stop, s.Size, s.ExitCode, s.Stacktrace, serverID)
	if err != nil {
		log.Printf("Error updating statistics: %s\n", err)
	}
	return err
}
