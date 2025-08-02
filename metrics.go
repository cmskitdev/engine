package engine

import (
	"sync"
	"time"
)

// Metrics tracks processing metrics.
type Metrics struct {
	ItemsProcessed        int64         `json:"items_processed"`
	ItemsSuccessful       int64         `json:"items_successful"`
	ItemsFailed           int64         `json:"items_failed"`
	TotalProcessingTime   time.Duration `json:"total_processing_time"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	StartedAt             time.Time     `json:"started_at"`
	LastProcessedAt       *time.Time    `json:"last_processed_at,omitempty"`
	mu                    sync.RWMutex
}
