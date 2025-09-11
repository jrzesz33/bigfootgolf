package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// AlertManager handles alerting and monitoring for the BOAT agent
type AlertManager struct {
	logger *logrus.Logger
	config *AlertConfig
}

// AlertConfig contains configuration for alerting thresholds
type AlertConfig struct {
	CostThreshold       float64       `json:"cost_threshold"`
	ErrorRateThreshold  float64       `json:"error_rate_threshold"`
	DurationThreshold   time.Duration `json:"duration_threshold"`
	ConsecutiveFailures int           `json:"consecutive_failures"`
}

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "INFO"
	AlertLevelWarning  AlertLevel = "WARNING"
	AlertLevelCritical AlertLevel = "CRITICAL"
)

// Alert represents an alert event
type Alert struct {
	ID          string                 `json:"id"`
	Level       AlertLevel             `json:"level"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	TaskID      string                 `json:"task_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewAlertManager creates a new alert manager
func NewAlertManager(config *AlertConfig) *AlertManager {
	if config == nil {
		config = &AlertConfig{
			CostThreshold:       100.0, // $100 threshold
			ErrorRateThreshold:  0.20,  // 20% error rate
			DurationThreshold:   time.Minute * 10, // 10 minute threshold
			ConsecutiveFailures: 3,     // 3 consecutive failures
		}
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	return &AlertManager{
		logger: logger,
		config: config,
	}
}

// CheckCostThreshold checks if deployment cost exceeds threshold
func (a *AlertManager) CheckCostThreshold(taskID string, estimatedCost float64) *Alert {
	if estimatedCost > a.config.CostThreshold {
		alert := &Alert{
			ID:          fmt.Sprintf("cost-threshold-%s-%d", taskID, time.Now().Unix()),
			Level:       AlertLevelWarning,
			Title:       "Cost Threshold Exceeded",
			Description: fmt.Sprintf("Estimated cost ($%.2f) exceeds threshold ($%.2f)", estimatedCost, a.config.CostThreshold),
			TaskID:      taskID,
			Timestamp:   time.Now().UTC(),
			Metadata: map[string]interface{}{
				"estimated_cost": estimatedCost,
				"threshold":      a.config.CostThreshold,
				"overage":        estimatedCost - a.config.CostThreshold,
			},
		}
		
		a.fireAlert(alert)
		return alert
	}
	return nil
}

// CheckTaskDuration checks if task duration exceeds threshold
func (a *AlertManager) CheckTaskDuration(taskID string, duration time.Duration) *Alert {
	if duration > a.config.DurationThreshold {
		alert := &Alert{
			ID:          fmt.Sprintf("duration-threshold-%s-%d", taskID, time.Now().Unix()),
			Level:       AlertLevelWarning,
			Title:       "Task Duration Threshold Exceeded",
			Description: fmt.Sprintf("Task duration (%s) exceeds threshold (%s)", duration, a.config.DurationThreshold),
			TaskID:      taskID,
			Timestamp:   time.Now().UTC(),
			Metadata: map[string]interface{}{
				"duration_ms":   duration.Milliseconds(),
				"threshold_ms":  a.config.DurationThreshold.Milliseconds(),
				"overage_ms":    (duration - a.config.DurationThreshold).Milliseconds(),
			},
		}
		
		a.fireAlert(alert)
		return alert
	}
	return nil
}

// CheckErrorRate checks if error rate exceeds threshold
func (a *AlertManager) CheckErrorRate(successfulTasks, failedTasks int) *Alert {
	totalTasks := successfulTasks + failedTasks
	if totalTasks == 0 {
		return nil
	}

	errorRate := float64(failedTasks) / float64(totalTasks)
	if errorRate > a.config.ErrorRateThreshold {
		alert := &Alert{
			ID:          fmt.Sprintf("error-rate-%d", time.Now().Unix()),
			Level:       AlertLevelCritical,
			Title:       "High Error Rate Detected",
			Description: fmt.Sprintf("Error rate (%.2f%%) exceeds threshold (%.2f%%)", errorRate*100, a.config.ErrorRateThreshold*100),
			Timestamp:   time.Now().UTC(),
			Metadata: map[string]interface{}{
				"error_rate":     errorRate,
				"threshold":      a.config.ErrorRateThreshold,
				"failed_tasks":   failedTasks,
				"successful_tasks": successfulTasks,
				"total_tasks":    totalTasks,
			},
		}
		
		a.fireAlert(alert)
		return alert
	}
	return nil
}

// CheckConsecutiveFailures tracks consecutive task failures
func (a *AlertManager) CheckConsecutiveFailures(failures int, lastFailedTaskIDs []string) *Alert {
	if failures >= a.config.ConsecutiveFailures {
		alert := &Alert{
			ID:          fmt.Sprintf("consecutive-failures-%d", time.Now().Unix()),
			Level:       AlertLevelCritical,
			Title:       "Consecutive Task Failures",
			Description: fmt.Sprintf("Detected %d consecutive task failures", failures),
			Timestamp:   time.Now().UTC(),
			Metadata: map[string]interface{}{
				"failure_count":     failures,
				"threshold":         a.config.ConsecutiveFailures,
				"failed_task_ids":   lastFailedTaskIDs,
			},
		}
		
		a.fireAlert(alert)
		return alert
	}
	return nil
}

// CheckMissingRequirements alerts when critical information is missing
func (a *AlertManager) CheckMissingRequirements(taskID string, missingFields []string) *Alert {
	if len(missingFields) > 0 {
		alert := &Alert{
			ID:          fmt.Sprintf("missing-requirements-%s-%d", taskID, time.Now().Unix()),
			Level:       AlertLevelWarning,
			Title:       "Missing Container Requirements",
			Description: fmt.Sprintf("Task requires additional information: %v", missingFields),
			TaskID:      taskID,
			Timestamp:   time.Now().UTC(),
			Metadata: map[string]interface{}{
				"missing_fields": missingFields,
				"field_count":    len(missingFields),
			},
		}
		
		a.fireAlert(alert)
		return alert
	}
	return nil
}

// CheckFreeTierCompliance alerts if deployment might exceed AWS free tier
func (a *AlertManager) CheckFreeTierCompliance(taskID string, resourceCounts map[string]int, estimatedCost float64) *Alert {
	violations := []string{}
	
	// Check resource limits
	freeTierLimits := map[string]int{
		"AWS::ECS::Service": 10,
		"AWS::ElasticLoadBalancingV2::LoadBalancer": 1,
		"AWS::EC2::SecurityGroup": 20,
	}

	for resourceType, count := range resourceCounts {
		if limit, exists := freeTierLimits[resourceType]; exists && count > limit {
			violations = append(violations, fmt.Sprintf("%s: %d (limit: %d)", resourceType, count, limit))
		}
	}

	// Check cost limit
	if estimatedCost > 25.0 { // Conservative free tier estimate
		violations = append(violations, fmt.Sprintf("Estimated cost: $%.2f (recommended limit: $25.00)", estimatedCost))
	}

	if len(violations) > 0 {
		alert := &Alert{
			ID:          fmt.Sprintf("free-tier-compliance-%s-%d", taskID, time.Now().Unix()),
			Level:       AlertLevelWarning,
			Title:       "Free Tier Compliance Warning",
			Description: fmt.Sprintf("Deployment may exceed AWS free tier limits: %v", violations),
			TaskID:      taskID,
			Timestamp:   time.Now().UTC(),
			Metadata: map[string]interface{}{
				"violations":       violations,
				"resource_counts":  resourceCounts,
				"estimated_cost":   estimatedCost,
			},
		}
		
		a.fireAlert(alert)
		return alert
	}
	return nil
}

// MonitorHealthCheck monitors application health and performance
func (a *AlertManager) MonitorHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.performHealthCheck()
		}
	}
}

// performHealthCheck performs periodic health checks
func (a *AlertManager) performHealthCheck() {
	healthStatus := map[string]interface{}{
		"timestamp":    time.Now().UTC(),
		"status":       "healthy",
		"uptime":       time.Since(startTime), // Would need to track start time
		"memory_usage": "unknown", // Would need actual memory monitoring
		"goroutines":   "unknown", // Would need goroutine counting
	}

	a.logger.WithFields(logrus.Fields{
		"event_type":    "health_check",
		"health_status": healthStatus,
	}).Info("BOAT Agent health check")
}

// fireAlert processes and logs alerts
func (a *AlertManager) fireAlert(alert *Alert) {
	logLevel := logrus.InfoLevel
	switch alert.Level {
	case AlertLevelWarning:
		logLevel = logrus.WarnLevel
	case AlertLevelCritical:
		logLevel = logrus.ErrorLevel
	}

	a.logger.WithFields(logrus.Fields{
		"event_type":    "alert_fired",
		"alert_id":      alert.ID,
		"alert_level":   alert.Level,
		"alert_title":   alert.Title,
		"alert_desc":    alert.Description,
		"task_id":       alert.TaskID,
		"alert_metadata": alert.Metadata,
		"timestamp":     alert.Timestamp,
	}).Log(logLevel, "BOAT Agent alert fired")

	// In a production environment, this would:
	// - Send to SNS topic for email/SMS notifications
	// - Post to Slack/Teams channels
	// - Create CloudWatch alarms
	// - Send to external monitoring systems
}

// GetAlertSummary returns a summary of recent alerts
func (a *AlertManager) GetAlertSummary(since time.Time) map[string]int {
	// This would typically query a persistent store
	// For now, return a placeholder summary
	return map[string]int{
		"total_alerts":    0,
		"info_alerts":     0,
		"warning_alerts":  0,
		"critical_alerts": 0,
	}
}

// Global variable to track start time (would be better managed)
var startTime = time.Now()