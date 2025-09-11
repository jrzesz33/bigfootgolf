package monitoring

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// MetricsCollector handles application metrics collection
type MetricsCollector struct {
	logger *logrus.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	return &MetricsCollector{
		logger: logger,
	}
}

// TaskMetrics represents metrics for a task execution
type TaskMetrics struct {
	TaskID           string        `json:"task_id"`
	TaskType         string        `json:"task_type"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	Duration         time.Duration `json:"duration_ms"`
	Status           string        `json:"status"`
	ResourcesCreated int           `json:"resources_created"`
	EstimatedCost    float64       `json:"estimated_cost"`
	ErrorCount       int           `json:"error_count"`
	WarningCount     int           `json:"warning_count"`
	ClaudeAPICalls   int           `json:"claude_api_calls"`
	AWSAPICalls      int           `json:"aws_api_calls"`
}

// LogTaskStart logs the start of a task
func (m *MetricsCollector) LogTaskStart(taskID, taskType string) {
	m.logger.WithFields(logrus.Fields{
		"event_type": "task_started",
		"task_id":    taskID,
		"task_type":  taskType,
		"timestamp":  time.Now().UTC(),
	}).Info("Task execution started")
}

// LogTaskCompletion logs task completion with comprehensive metrics
func (m *MetricsCollector) LogTaskCompletion(metrics *TaskMetrics) {
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	m.logger.WithFields(logrus.Fields{
		"event_type":        "task_completed",
		"task_id":           metrics.TaskID,
		"task_type":         metrics.TaskType,
		"duration_ms":       metrics.Duration.Milliseconds(),
		"status":            metrics.Status,
		"resources_created": metrics.ResourcesCreated,
		"estimated_cost":    metrics.EstimatedCost,
		"error_count":       metrics.ErrorCount,
		"warning_count":     metrics.WarningCount,
		"claude_api_calls":  metrics.ClaudeAPICalls,
		"aws_api_calls":     metrics.AWSAPICalls,
		"cost_per_resource": m.calculateCostPerResource(metrics),
		"success_rate":      m.calculateSuccessRate(metrics),
	}).Info("Task execution completed")
}

// LogAWSResourceCreation logs AWS resource creation events
func (m *MetricsCollector) LogAWSResourceCreation(taskID, resourceType, resourceID string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}

	m.logger.WithFields(logrus.Fields{
		"event_type":    "aws_resource_created",
		"task_id":       taskID,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"duration_ms":   duration.Milliseconds(),
		"status":        status,
		"timestamp":     time.Now().UTC(),
	}).Info("AWS resource creation event")
}

// LogClaudeAPICall logs Claude API interaction metrics
func (m *MetricsCollector) LogClaudeAPICall(taskID string, promptTokens, completionTokens int, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}

	m.logger.WithFields(logrus.Fields{
		"event_type":         "claude_api_call",
		"task_id":            taskID,
		"prompt_tokens":      promptTokens,
		"completion_tokens":  completionTokens,
		"total_tokens":       promptTokens + completionTokens,
		"duration_ms":        duration.Milliseconds(),
		"status":             status,
		"tokens_per_second":  float64(promptTokens+completionTokens) / duration.Seconds(),
		"timestamp":          time.Now().UTC(),
	}).Info("Claude API call metrics")
}

// LogCostWarning logs cost-related warnings
func (m *MetricsCollector) LogCostWarning(taskID string, estimatedCost, threshold float64, resourceType string) {
	m.logger.WithFields(logrus.Fields{
		"event_type":     "cost_warning",
		"task_id":        taskID,
		"estimated_cost": estimatedCost,
		"threshold":      threshold,
		"resource_type":  resourceType,
		"overage":        estimatedCost - threshold,
		"overage_pct":    ((estimatedCost - threshold) / threshold) * 100,
		"timestamp":      time.Now().UTC(),
	}).Warn("Cost threshold exceeded")
}

// LogValidationFailure logs validation failure events
func (m *MetricsCollector) LogValidationFailure(taskID string, validationType string, missingFields []string, errors []string) {
	m.logger.WithFields(logrus.Fields{
		"event_type":       "validation_failure",
		"task_id":          taskID,
		"validation_type":  validationType,
		"missing_fields":   missingFields,
		"error_count":      len(errors),
		"errors":           errors,
		"timestamp":        time.Now().UTC(),
	}).Error("Validation failure")
}

// LogPerformanceMetrics logs overall system performance metrics
func (m *MetricsCollector) LogPerformanceMetrics(ctx context.Context) {
	// This could collect system-wide metrics periodically
	m.logger.WithFields(logrus.Fields{
		"event_type": "performance_snapshot",
		"timestamp":  time.Now().UTC(),
		// Could include memory usage, goroutine count, etc.
	}).Info("Performance metrics snapshot")
}

// Helper methods

func (m *MetricsCollector) calculateCostPerResource(metrics *TaskMetrics) float64 {
	if metrics.ResourcesCreated == 0 {
		return 0
	}
	return metrics.EstimatedCost / float64(metrics.ResourcesCreated)
}

func (m *MetricsCollector) calculateSuccessRate(metrics *TaskMetrics) float64 {
	totalOperations := metrics.ResourcesCreated + metrics.ErrorCount
	if totalOperations == 0 {
		return 0
	}
	return float64(metrics.ResourcesCreated) / float64(totalOperations) * 100
}

// AggregatedMetrics represents aggregated metrics over time
type AggregatedMetrics struct {
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	TotalTasks      int       `json:"total_tasks"`
	SuccessfulTasks int       `json:"successful_tasks"`
	FailedTasks     int       `json:"failed_tasks"`
	SuccessRate     float64   `json:"success_rate"`
	AverageDuration time.Duration `json:"average_duration_ms"`
	TotalCost       float64   `json:"total_cost"`
	AverageCost     float64   `json:"average_cost"`
	ResourcesCreated int      `json:"total_resources_created"`
	MostCommonTaskType string `json:"most_common_task_type"`
	TotalClaudeAPICalls int   `json:"total_claude_api_calls"`
	TotalAWSAPICalls    int   `json:"total_aws_api_calls"`
}

// LogAggregatedMetrics logs periodic aggregated metrics
func (m *MetricsCollector) LogAggregatedMetrics(metrics *AggregatedMetrics) {
	m.logger.WithFields(logrus.Fields{
		"event_type":             "aggregated_metrics",
		"period_start":           metrics.PeriodStart,
		"period_end":             metrics.PeriodEnd,
		"total_tasks":            metrics.TotalTasks,
		"successful_tasks":       metrics.SuccessfulTasks,
		"failed_tasks":           metrics.FailedTasks,
		"success_rate":           metrics.SuccessRate,
		"average_duration_ms":    metrics.AverageDuration.Milliseconds(),
		"total_cost":             metrics.TotalCost,
		"average_cost_per_task":  metrics.AverageCost,
		"resources_created":      metrics.ResourcesCreated,
		"most_common_task_type":  metrics.MostCommonTaskType,
		"total_claude_api_calls": metrics.TotalClaudeAPICalls,
		"total_aws_api_calls":    metrics.TotalAWSAPICalls,
		"cost_per_resource":      metrics.TotalCost / float64(max(metrics.ResourcesCreated, 1)),
		"timestamp":              time.Now().UTC(),
	}).Info("Aggregated metrics report")
}

// max is a helper function since Go doesn't have a built-in max for int
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}