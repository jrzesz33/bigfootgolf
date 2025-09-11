package deployment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"bigfoot/golf/opsagent/internal/aws"
	"bigfoot/golf/opsagent/internal/models"
)

// Manager handles deployment operations
type Manager struct {
	awsManager *aws.Manager
}

// NewManager creates a new deployment manager
func NewManager(awsManager *aws.Manager) *Manager {
	return &Manager{
		awsManager: awsManager,
	}
}

// ExecutePlan executes a deployment plan
func (m *Manager) ExecutePlan(ctx context.Context, plan *models.DeploymentPlan) (*models.TaskResult, error) {
	logrus.WithFields(logrus.Fields{
		"task_id":         plan.TaskID,
		"resource_count":  len(plan.Resources),
		"estimated_cost":  plan.EstimatedCost,
	}).Info("Starting deployment plan execution")

	result := &models.TaskResult{
		TaskID:            plan.TaskID,
		DeployedResources: []string{},
		Warnings:          plan.Warnings,
		NextSteps:         []string{},
		Metadata:          make(map[string]interface{}),
	}

	// Check prerequisites
	if len(plan.Prerequisites) > 0 {
		logrus.WithField("prerequisites", plan.Prerequisites).Info("Checking prerequisites")
		for _, prereq := range plan.Prerequisites {
			if err := m.checkPrerequisite(ctx, prereq); err != nil {
				result.Status = "failed"
				result.Message = fmt.Sprintf("Prerequisite check failed: %s - %v", prereq, err)
				return result, err
			}
		}
		logrus.Info("All prerequisites satisfied")
	}

	// Execute resources in dependency order
	deployedResources := make(map[string]string) // resource name -> AWS ID mapping
	
	for i, resource := range plan.Resources {
		logrus.WithFields(logrus.Fields{
			"resource_name": resource.Name,
			"resource_type": resource.Type,
			"step":          fmt.Sprintf("%d/%d", i+1, len(plan.Resources)),
		}).Info("Deploying resource")

		// Check dependencies
		if err := m.checkDependencies(resource.DependsOn, deployedResources); err != nil {
			result.Status = "failed"
			result.Message = fmt.Sprintf("Dependency check failed for %s: %v", resource.Name, err)
			return result, err
		}

		// Deploy the resource
		resourceID, err := m.deployResource(ctx, &resource)
		if err != nil {
			result.Status = "failed"
			result.Message = fmt.Sprintf("Failed to deploy resource %s: %v", resource.Name, err)
			result.Warnings = append(result.Warnings, fmt.Sprintf("Partial deployment completed: %d/%d resources", i, len(plan.Resources)))
			return result, err
		}

		deployedResources[resource.Name] = resourceID
		result.DeployedResources = append(result.DeployedResources, fmt.Sprintf("%s (%s)", resource.Name, resourceID))
		
		logrus.WithFields(logrus.Fields{
			"resource_name": resource.Name,
			"resource_id":   resourceID,
		}).Info("Resource deployed successfully")
	}

	// Generate next steps
	result.NextSteps = m.generateNextSteps(plan, deployedResources)
	result.Status = "success"
	result.Message = fmt.Sprintf("Successfully deployed %d resources", len(plan.Resources))
	
	// Add metadata
	result.Metadata["deployment_time"] = time.Now().UTC().Format(time.RFC3339)
	result.Metadata["estimated_cost"] = plan.EstimatedCost
	result.Metadata["resource_mapping"] = deployedResources

	logrus.WithFields(logrus.Fields{
		"task_id":        plan.TaskID,
		"deployed_count": len(result.DeployedResources),
		"duration":       "calculated", // Would track actual duration
	}).Info("Deployment plan execution completed successfully")

	return result, nil
}

// deployResource deploys a single AWS resource
func (m *Manager) deployResource(ctx context.Context, resource *models.AWSResourceDefinition) (string, error) {
	switch resource.Type {
	case "AWS::ECS::Service":
		return m.deployECSService(ctx, resource)
	case "AWS::ECS::TaskDefinition":
		return m.deployTaskDefinition(ctx, resource)
	case "AWS::ElasticLoadBalancingV2::LoadBalancer":
		return m.deployLoadBalancer(ctx, resource)
	case "AWS::EC2::SecurityGroup":
		return m.deploySecurityGroup(ctx, resource)
	default:
		// Use Cloud Control API for other resource types
		return m.awsManager.CreateResource(ctx, resource.Type, resource.Properties)
	}
}

// deployECSService deploys an ECS service with optimized settings
func (m *Manager) deployECSService(ctx context.Context, resource *models.AWSResourceDefinition) (string, error) {
	// Get VPC info for networking
	vpcInfo, err := m.awsManager.GetVPCInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get VPC info: %w", err)
	}

	// Optimize properties for cost
	properties := resource.Properties
	if properties == nil {
		properties = make(map[string]interface{})
	}

	// Set cost-optimized defaults
	properties["LaunchType"] = "FARGATE" // Fargate for minimal management overhead
	properties["DesiredCount"] = 1       // Start with minimal instances
	properties["NetworkConfiguration"] = map[string]interface{}{
		"AwsvpcConfiguration": map[string]interface{}{
			"Subnets": vpcInfo.Subnets,
			"SecurityGroups": []string{}, // Will be populated by security group deployment
			"AssignPublicIp": "ENABLED",  // For free tier networking
		},
	}

	return m.awsManager.CreateResource(ctx, resource.Type, properties)
}

// deployTaskDefinition deploys an ECS task definition
func (m *Manager) deployTaskDefinition(ctx context.Context, resource *models.AWSResourceDefinition) (string, error) {
	properties := resource.Properties
	if properties == nil {
		properties = make(map[string]interface{})
	}

	// Set cost-optimized defaults for Fargate
	properties["RequiresCompatibilities"] = []string{"FARGATE"}
	properties["NetworkMode"] = "awsvpc"
	properties["Cpu"] = "256"    // Minimum Fargate CPU
	properties["Memory"] = "512" // Minimum Fargate memory

	return m.awsManager.CreateResource(ctx, resource.Type, properties)
}

// deployLoadBalancer deploys an Application Load Balancer
func (m *Manager) deployLoadBalancer(ctx context.Context, resource *models.AWSResourceDefinition) (string, error) {
	vpcInfo, err := m.awsManager.GetVPCInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get VPC info: %w", err)
	}

	properties := resource.Properties
	if properties == nil {
		properties = make(map[string]interface{})
	}

	// Configure for free tier
	properties["Type"] = "application"
	properties["Scheme"] = "internet-facing"
	properties["IpAddressType"] = "ipv4"
	properties["Subnets"] = vpcInfo.Subnets

	return m.awsManager.CreateResource(ctx, resource.Type, properties)
}

// deploySecurityGroup deploys a security group
func (m *Manager) deploySecurityGroup(ctx context.Context, resource *models.AWSResourceDefinition) (string, error) {
	vpcInfo, err := m.awsManager.GetVPCInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get VPC info: %w", err)
	}

	properties := resource.Properties
	if properties == nil {
		properties = make(map[string]interface{})
	}

	properties["VpcId"] = vpcInfo.VpcID
	properties["GroupDescription"] = "BOAT Agent managed security group"

	return m.awsManager.CreateResource(ctx, resource.Type, properties)
}

// checkPrerequisite checks if a prerequisite is satisfied
func (m *Manager) checkPrerequisite(ctx context.Context, prerequisite string) error {
	// This would implement specific prerequisite checks
	// For example: checking if ECS cluster exists, IAM roles are configured, etc.
	logrus.WithField("prerequisite", prerequisite).Info("Checking prerequisite (placeholder)")
	return nil
}

// checkDependencies verifies that all dependencies are deployed
func (m *Manager) checkDependencies(dependencies []string, deployedResources map[string]string) error {
	for _, dep := range dependencies {
		if _, exists := deployedResources[dep]; !exists {
			return fmt.Errorf("dependency %s not found in deployed resources", dep)
		}
	}
	return nil
}

// generateNextSteps creates actionable next steps for the user
func (m *Manager) generateNextSteps(plan *models.DeploymentPlan, deployedResources map[string]string) []string {
	nextSteps := []string{
		"Verify deployed resources are healthy and running",
		"Configure monitoring and logging for the deployed services",
		"Update DNS records if public-facing services were deployed",
		"Review security group rules and ensure least privilege access",
		"Set up automated backups if applicable",
	}

	// Add specific steps based on deployed resources
	for resourceName, resourceID := range deployedResources {
		if strings.Contains(resourceName, "LoadBalancer") {
			nextSteps = append(nextSteps, fmt.Sprintf("Configure health checks for load balancer %s", resourceID))
		}
		if strings.Contains(resourceName, "ECS") {
			nextSteps = append(nextSteps, fmt.Sprintf("Monitor ECS service %s for proper scaling", resourceID))
		}
	}

	return nextSteps
}

