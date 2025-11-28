package services

import "github.com/tupyy/assisted-migration-agent/pkg/scheduler"

type Collector struct {
	scheduler *scheduler.Scheduler
}

func NewCollectorService(s *scheduler.Scheduler) *Collector {
	return &Collector{scheduler: s}
}
