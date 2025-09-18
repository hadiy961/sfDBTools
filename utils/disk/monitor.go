// Package disk provides disk monitoring functionality.
// This file contains background monitoring and alerting capabilities.
package disk

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/fs"
)

// MonitorConfig contains configuration for disk monitoring.
type MonitorConfig struct {
	Path             string                    // Path to monitor
	Interval         time.Duration             // Check interval
	ThresholdPercent float64                   // Used percentage threshold
	AlertCallback    func(usage *fs.DiskUsage) // Callback when threshold exceeded
}

// Monitor represents a disk monitor instance.
type Monitor struct {
	config  MonitorConfig
	stopCh  chan struct{}
	running bool
	lgr     *logger.Logger
}

// NewMonitor creates a new disk monitor with the given configuration.
func NewMonitor(config MonitorConfig) *Monitor {
	lg, _ := logger.Get()

	if config.Interval <= 0 {
		config.Interval = 30 * time.Second // Default 30 seconds
	}
	if config.ThresholdPercent <= 0 {
		config.ThresholdPercent = 90.0 // Default 90%
	}

	return &Monitor{
		config: config,
		stopCh: make(chan struct{}),
		lgr:    lg,
	}
}

// Start begins monitoring disk usage in the background.
// Returns an error if the monitor is already running.
func (m *Monitor) Start() error {
	if m.running {
		return fmt.Errorf("monitor is already running")
	}

	m.running = true
	go m.monitorLoop()

	m.lgr.Info("Disk monitor started",
		logger.String("path", m.config.Path),
		logger.String("interval", m.config.Interval.String()),
		logger.Float64("threshold", m.config.ThresholdPercent))

	return nil
}

// Stop stops the disk monitor.
func (m *Monitor) Stop() {
	if !m.running {
		return
	}

	close(m.stopCh)
	m.running = false

	m.lgr.Info("Disk monitor stopped",
		logger.String("path", m.config.Path))
}

// IsRunning returns true if the monitor is currently running.
func (m *Monitor) IsRunning() bool {
	return m.running
}

// monitorLoop is the main monitoring loop that runs in a goroutine.
func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	fsManager := fs.NewManager()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			usage, err := fsManager.Dir().GetDiskUsage(m.config.Path)
			if err != nil {
				m.lgr.Error("Failed to get disk usage in monitor",
					logger.String("path", m.config.Path),
					logger.Error(err))
				continue
			}

			if usage.UsedPercent >= m.config.ThresholdPercent {
				m.lgr.Warn("Disk usage threshold exceeded",
					logger.String("path", m.config.Path),
					logger.Float64("used_percent", usage.UsedPercent),
					logger.Float64("threshold", m.config.ThresholdPercent))

				if m.config.AlertCallback != nil {
					m.config.AlertCallback(usage)
				}
			} else {
				m.lgr.Debug("Disk usage check passed",
					logger.String("path", m.config.Path),
					logger.Float64("used_percent", usage.UsedPercent))
			}
		}
	}
}

// MonitorDisk provides a simple function-based monitoring interface.
// It starts monitoring and returns a stop function.
// This function maintains backward compatibility with the original API.
func MonitorDisk(path string, interval time.Duration, thresholdPercent float64, cb func(*fs.DiskUsage)) func() {
	config := MonitorConfig{
		Path:             path,
		Interval:         interval,
		ThresholdPercent: thresholdPercent,
		AlertCallback:    cb,
	}

	monitor := NewMonitor(config)
	if err := monitor.Start(); err != nil {
		lg, _ := logger.Get()
		lg.Error("Failed to start disk monitor", logger.Error(err))
		return func() {} // Return no-op function
	}

	// Return stop function
	return monitor.Stop
}
