package main

import (
	"time"

	"github.com/3Blades/go-sdk/client/projects"
)

// Stats is base statistics gatherer
type Stats struct {
	Start      time.Time `json:"start"`
	Stop       time.Time `json:"stop,omitempty"`
	Size       int64     `json:"size"`
	ExitCode   int       `json:"exit_code"`
	Stacktrace string    `json:"stacktrace"`
}

// NewStats creates Stats object
func NewStats() *Stats {
	return &Stats{Size: 0}
}

// Duration calculates run duration
func (s *Stats) Duration() time.Duration {
	return s.Stop.Sub(s.Start) / time.Millisecond
}

// Send writes statistics data to database
func (s *Stats) Send(args *Args) error {
	cli := APIClient(args.ApiRoot, args.ApiKey)
	params := projects.NewProjectsServersRunStatsCreateParams()
	params.SetNamespace(args.Namespace)
	params.SetProjectPk(args.ProjectID)
	params.SetServerPk(args.ServerID)
	params.SetData(projects.ProjectsServersRunStatsCreateBody{
		Start:      s.Start.Format(time.RFC3339),
		Stop:       s.Stop.Format(time.RFC3339),
		ExitCode:   int64(s.ExitCode),
		Stacktrace: s.Stacktrace,
	})
	_, err := cli.Projects.ProjectsServersRunStatsCreate(params)
	return err
}
