
package overlays

import (
	"context"
	"time"

	"github.com/ubom/workflow/layer0"
	"github.com/ubom/workflow/executors"
)

// ExecutorOverlay defines the interface for executor overlays
type ExecutorOverlay interface {
	executors.WorkExecutor
	SetNext(executor executors.WorkExecutor)
}

// BaseOverlay provides common overlay functionality
type BaseOverlay struct {
	next executors.WorkExecutor
}

// SetNext sets the next executor in the chain
func (o *BaseOverlay) SetNext(executor executors.WorkExecutor) {
	o.next = executor
}

// CanExecute delegates to the next executor
func (o *BaseOverlay) CanExecute(workType layer0.WorkType) bool {
	if o.next == nil {
		return false
	}
	return o.next.CanExecute(workType)
}

// GetSupportedTypes delegates to the next executor
func (o *BaseOverlay) GetSupportedTypes() []layer0.WorkType {
	if o.next == nil {
		return []layer0.WorkType{}
	}
	return o.next.GetSupportedTypes()
}

// Validate delegates to the next executor
func (o *BaseOverlay) Validate(work layer0.Work) error {
	if o.next == nil {
		return nil
	}
	return o.next.Validate(work)
}

// GetSchema delegates to the next executor
func (o *BaseOverlay) GetSchema() executors.WorkSchema {
	if o.next == nil {
		return executors.WorkSchema{}
	}
	return o.next.GetSchema()
}

// GetMetadata delegates to the next executor
func (o *BaseOverlay) GetMetadata() executors.WorkMetadata {
	if o.next == nil {
		return executors.WorkMetadata{}
	}
	return o.next.GetMetadata()
}

// MetricsOverlay adds metrics collection to execution
type MetricsOverlay struct {
	BaseOverlay
	metricsCollector MetricsCollector
}

// MetricsCollector interface for collecting metrics
type MetricsCollector interface {
	RecordExecution(workType layer0.WorkType, duration time.Duration, success bool)
	RecordResourceUsage(cpu float64, memory int64)
}

// NewMetricsOverlay creates a new metrics overlay
func NewMetricsOverlay(collector MetricsCollector) *MetricsOverlay {
	return &MetricsOverlay{
		metricsCollector: collector,
	}
}

// Execute adds metrics collection around execution
func (m *MetricsOverlay) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
	if m.next == nil {
		return executors.WorkResult{
			Success: false,
			Error:   "no executor configured",
		}, nil
	}

	startTime := time.Now()
	result, err := m.next.Execute(ctx, work, workContext)
	duration := time.Since(startTime)

	// Record metrics
	if m.metricsCollector != nil {
		m.metricsCollector.RecordExecution(work.GetType(), duration, result.Success)
		if result.Metrics.CPUUsage > 0 || result.Metrics.MemoryUsage > 0 {
			m.metricsCollector.RecordResourceUsage(result.Metrics.CPUUsage, result.Metrics.MemoryUsage)
		}
	}

	return result, err
}

// LoggingOverlay adds logging to execution
type LoggingOverlay struct {
	BaseOverlay
	logger Logger
}

// Logger interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewLoggingOverlay creates a new logging overlay
func NewLoggingOverlay(logger Logger) *LoggingOverlay {
	return &LoggingOverlay{
		logger: logger,
	}
}

// Execute adds logging around execution
func (l *LoggingOverlay) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
	if l.next == nil {
		return executors.WorkResult{
			Success: false,
			Error:   "no executor configured",
		}, nil
	}

	if l.logger != nil {
		l.logger.Info("Starting work execution", "workID", work.GetID(), "workType", work.GetType())
	}

	result, err := l.next.Execute(ctx, work, workContext)

	if l.logger != nil {
		if result.Success {
			l.logger.Info("Work execution completed successfully", 
				"workID", work.GetID(), 
				"duration", result.Metrics.Duration)
		} else {
			l.logger.Error("Work execution failed", 
				"workID", work.GetID(), 
				"error", result.Error)
		}
	}

	return result, err
}

// RetryOverlay adds retry functionality to execution
type RetryOverlay struct {
	BaseOverlay
	maxAttempts int
	delay       time.Duration
	backoff     string // "linear" or "exponential"
}

// NewRetryOverlay creates a new retry overlay
func NewRetryOverlay(maxAttempts int, delay time.Duration, backoff string) *RetryOverlay {
	return &RetryOverlay{
		maxAttempts: maxAttempts,
		delay:       delay,
		backoff:     backoff,
	}
}

// Execute adds retry logic around execution
func (r *RetryOverlay) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
	if r.next == nil {
		return executors.WorkResult{
			Success: false,
			Error:   "no executor configured",
		}, nil
	}

	var lastResult executors.WorkResult
	var lastErr error

	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		result, err := r.next.Execute(ctx, work, workContext)
		
		if result.Success {
			return result, err
		}

		lastResult = result
		lastErr = err

		// Don't wait after the last attempt
		if attempt < r.maxAttempts {
			delay := r.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return lastResult, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return lastResult, lastErr
}

// calculateDelay calculates the delay for the given attempt
func (r *RetryOverlay) calculateDelay(attempt int) time.Duration {
	switch r.backoff {
	case "exponential":
		return r.delay * time.Duration(1<<uint(attempt-1))
	default: // linear
		return r.delay * time.Duration(attempt)
	}
}
