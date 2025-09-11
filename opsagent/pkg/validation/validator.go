package validation

import (
	"fmt"
	"regexp"
	"strings"

	"bigfoot/golf/opsagent/internal/models"
)

// Validator handles container requirement validation
type Validator struct {
	registryPattern *regexp.Regexp
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	// Pattern to validate container registry URLs
	registryPattern := regexp.MustCompile(`^ghcr\.io/jrzesz33/[a-z0-9\-_/]+:[a-z0-9\-_.]+$`)
	
	return &Validator{
		registryPattern: registryPattern,
	}
}

// ValidateContainerRequirements validates all container requirements
func (v *Validator) ValidateContainerRequirements(req *models.ContainerRequirements) error {
	var missingFields []string
	var validationNotes []string

	// Validate container image
	if req.Image == "" {
		missingFields = append(missingFields, "container_image")
	} else if !v.registryPattern.MatchString(req.Image) {
		validationNotes = append(validationNotes, "Container image must be from ghcr.io/jrzesz33/ registry")
		missingFields = append(missingFields, "valid_container_image")
	}

	// Validate ports
	if len(req.Ports) == 0 {
		missingFields = append(missingFields, "port_mappings")
		validationNotes = append(validationNotes, "At least one port mapping is required")
	} else {
		for _, port := range req.Ports {
			if port.ContainerPort <= 0 || port.ContainerPort > 65535 {
				validationNotes = append(validationNotes, fmt.Sprintf("Invalid container port: %d", port.ContainerPort))
			}
			if port.Protocol == "" {
				port.Protocol = "tcp" // Default to TCP
			}
		}
	}

	// Validate health check
	if !req.HealthCheck.Enabled {
		validationNotes = append(validationNotes, "Health check is disabled - consider enabling for production deployments")
	} else {
		if req.HealthCheck.Path == "" && req.HealthCheck.Protocol == "http" {
			missingFields = append(missingFields, "health_check_path")
		}
		if req.HealthCheck.Port <= 0 {
			missingFields = append(missingFields, "health_check_port")
		}
		if req.HealthCheck.Protocol == "" {
			req.HealthCheck.Protocol = "http" // Default to HTTP
		}
	}

	// Validate external egress requirements
	if len(req.ExternalEgress) == 0 {
		validationNotes = append(validationNotes, "No external egress requirements specified - container will have limited internet access")
	} else {
		for _, egress := range req.ExternalEgress {
			if !v.isValidEgressURL(egress) {
				validationNotes = append(validationNotes, fmt.Sprintf("Invalid egress URL format: %s", egress))
			}
		}
	}

	// Check for environment variables
	if len(req.EnvironmentVariables) == 0 {
		validationNotes = append(validationNotes, "No environment variables specified")
	}

	// Validate secrets configuration
	if len(req.DynamicSecrets) == 0 && len(req.ExistingSecrets) == 0 {
		validationNotes = append(validationNotes, "No secrets configuration specified - consider if the application needs secrets")
	}

	// Update the requirements with validation results
	req.MissingFields = missingFields
	req.ValidationNotes = validationNotes
	req.IsValid = len(missingFields) == 0

	if !req.IsValid {
		return fmt.Errorf("container requirements validation failed: missing %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// ValidateDeploymentPlan validates a deployment plan for cost and compliance
func (v *Validator) ValidateDeploymentPlan(plan *models.DeploymentPlan, maxCost float64) error {
	if plan.EstimatedCost > maxCost {
		return fmt.Errorf("estimated cost %.2f exceeds maximum allowed cost %.2f", plan.EstimatedCost, maxCost)
	}

	// Validate resource types are within free tier limits
	resourceCounts := make(map[string]int)
	for _, resource := range plan.Resources {
		resourceCounts[resource.Type]++
	}

	// Check free tier limits
	if err := v.validateFreeTierLimits(resourceCounts); err != nil {
		return err
	}

	return nil
}

// isValidEgressURL validates external egress URL format
func (v *Validator) isValidEgressURL(url string) bool {
	// Basic validation for external URLs
	return strings.HasPrefix(url, "http://") || 
		   strings.HasPrefix(url, "https://") || 
		   v.isValidIPOrDomain(url)
}

// isValidIPOrDomain checks if the string is a valid IP or domain
func (v *Validator) isValidIPOrDomain(addr string) bool {
	// Simple check for IP addresses and domain names
	if strings.Contains(addr, ".") {
		return true // Basic domain/IP check
	}
	return false
}

// validateFreeTierLimits ensures deployment stays within AWS free tier limits
func (v *Validator) validateFreeTierLimits(resourceCounts map[string]int) error {
	freeTierLimits := map[string]int{
		"AWS::ECS::Service":      10,  // Conservative limit for ECS services
		"AWS::ElasticLoadBalancingV2::LoadBalancer": 1, // ALB free tier
		"AWS::EC2::SecurityGroup": 20, // Security groups limit
		"AWS::ECS::TaskDefinition": 50, // Task definitions
	}

	for resourceType, count := range resourceCounts {
		if limit, exists := freeTierLimits[resourceType]; exists {
			if count > limit {
				return fmt.Errorf("resource type %s count (%d) exceeds free tier limit (%d)", 
					resourceType, count, limit)
			}
		}
	}

	return nil
}

// GetValidationSummary returns a human-readable validation summary
func (v *Validator) GetValidationSummary(req *models.ContainerRequirements) string {
	if req.IsValid {
		return "✅ All container requirements validated successfully"
	}

	summary := "❌ Container requirements validation failed:\n"
	
	if len(req.MissingFields) > 0 {
		summary += fmt.Sprintf("Missing required fields: %s\n", strings.Join(req.MissingFields, ", "))
	}
	
	if len(req.ValidationNotes) > 0 {
		summary += "Additional notes:\n"
		for _, note := range req.ValidationNotes {
			summary += fmt.Sprintf("- %s\n", note)
		}
	}
	
	return summary
}