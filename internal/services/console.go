package services

import (
	"sync"
	"time"

	"github.com/tupyy/assisted-migration-agent/internal/models"
	"github.com/tupyy/assisted-migration-agent/pkg/scheduler"
)

type Console struct {
	url            string
	updateInterval time.Duration
	status         models.ConsoleStatus
	scheduler      *scheduler.Scheduler
	mu             sync.Mutex
}

func NewConnectedConsoleService(s *scheduler.Scheduler, consoleURL string, updateInterval time.Duration) *Console {
	defaultStatus := models.ConsoleStatus{
		Current: models.ConsoleStatusDisconnected,
		Target:  models.ConsoleStatusConnected,
	}
	return newConsoleService(s, consoleURL, updateInterval, defaultStatus)
}

func NewConsoleService(s *scheduler.Scheduler, consoleURL string, updateInterval time.Duration) *Console {
	defaultStatus := models.ConsoleStatus{
		Current: models.ConsoleStatusDisconnected,
		Target:  models.ConsoleStatusDisconnected,
	}
	return newConsoleService(s, consoleURL, updateInterval, defaultStatus)
}

func newConsoleService(s *scheduler.Scheduler, consoleURL string, updateInterval time.Duration, defaultStatus models.ConsoleStatus) *Console {
	return &Console{
		url:            consoleURL,
		updateInterval: updateInterval,
		scheduler:      s,
		status:         defaultStatus,
	}
}

func (c *Console) SetMode(mode models.AgentMode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch mode {
	case models.AgentModeConnected:
		c.status.Target = models.ConsoleStatusConnected
	case models.AgentModeDisconnected:
		c.status.Target = models.ConsoleStatusDisconnected
	}
}

func (c *Console) Status() models.ConsoleStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}
