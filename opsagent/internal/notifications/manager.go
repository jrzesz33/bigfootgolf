package notifications

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Manager handles notification operations
type Manager struct{}

// NewManager creates a new notification manager
func NewManager() *Manager {
	return &Manager{}
}

// NotifyMissingRequirements logs a notification about missing container requirements
func (m *Manager) NotifyMissingRequirements(taskID string, missingFields []string) {
	logrus.WithFields(logrus.Fields{
		"task_id":        taskID,
		"missing_fields": strings.Join(missingFields, ", "),
		"notification_type": "missing_requirements",
	}).Warn("BOAT Agent needs assistance - missing container requirements")

	// In future iterations, this could send to:
	// - Slack/Teams channels
	// - Email notifications
	// - SNS topics
	// - EventBridge events for downstream processing
}

// NotifyDeploymentStatus notifies about deployment status changes
func (m *Manager) NotifyDeploymentStatus(taskID string, status string, message string) {
	logrus.WithFields(logrus.Fields{
		"task_id": taskID,
		"status":  status,
		"message": message,
		"notification_type": "deployment_status",
	}).Info("Deployment status update")
}

// NotifyError logs error notifications
func (m *Manager) NotifyError(taskID string, error error, context string) {
	logrus.WithFields(logrus.Fields{
		"task_id": taskID,
		"error":   error.Error(),
		"context": context,
		"notification_type": "error",
	}).Error("BOAT Agent encountered an error")
}

// NotifyCostWarning logs cost-related warnings
func (m *Manager) NotifyCostWarning(taskID string, estimatedCost float64, threshold float64) {
	logrus.WithFields(logrus.Fields{
		"task_id":        taskID,
		"estimated_cost": estimatedCost,
		"threshold":      threshold,
		"notification_type": "cost_warning",
	}).Warn("Deployment may exceed cost threshold")
}

// NotifySuccess logs successful completion
func (m *Manager) NotifySuccess(taskID string, deployedResources []string) {
	logrus.WithFields(logrus.Fields{
		"task_id":            taskID,
		"deployed_resources": strings.Join(deployedResources, ", "),
		"resource_count":     len(deployedResources),
		"notification_type":  "success",
	}).Info("BOAT Agent successfully completed deployment")
}

// GetRequiredAssistanceMessage formats a message for human intervention
func (m *Manager) GetRequiredAssistanceMessage(taskID string, missingFields []string, context string) string {
	return fmt.Sprintf(`BOAT Agent requires assistance for task %s:

Missing Container Requirements:
%s

Context: %s

Please provide the missing information to proceed with deployment.

Required format:
- Ports: containerPort:hostPort (e.g., 8080:80)
- Dynamic Secrets: list of secrets to be created
- Existing Secrets: list of existing secret names/ARNs
- Environment Variables: key=value pairs
- External Egress: list of external URLs/IPs the container needs access to
- Public Ports: list of ports to expose via load balancer
- Health Check: path, port, protocol, intervals

Once provided, the BOAT Agent will automatically proceed with deployment.`,
		taskID,
		strings.Join(missingFields, "\n- "),
		context)
}