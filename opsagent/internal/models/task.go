package models

// TaskRequest represents an incoming task request from EventBridge
type TaskRequest struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // e.g., "deploy", "scale", "troubleshoot"
	Description string                 `json:"description"` // Natural language description
	Parameters  map[string]interface{} `json:"parameters"`  // Additional structured parameters
	Priority    string                 `json:"priority"`    // "low", "medium", "high", "urgent"
	Requester   string                 `json:"requester"`   // Who requested the task
}

// ContainerRequirements represents the requirements for containerized applications
type ContainerRequirements struct {
	// Required container configuration
	Image                string            `json:"image"`
	Ports                []PortMapping     `json:"ports"`
	DynamicSecrets       []string          `json:"dynamic_secrets"`
	ExistingSecrets      []string          `json:"existing_secrets"`
	EnvironmentVariables map[string]string `json:"environment_variables"`
	ExternalEgress       []string          `json:"external_egress"`
	PublicPorts          []int             `json:"public_ports"`
	HealthCheck          HealthCheckConfig `json:"health_check"`

	// Validation status
	IsValid        bool     `json:"is_valid"`
	MissingFields  []string `json:"missing_fields"`
	ValidationNotes []string `json:"validation_notes"`
}

// PortMapping represents port configuration for containers
type PortMapping struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port"`
	Protocol      string `json:"protocol"` // "tcp", "udp"
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Enabled  bool   `json:"enabled"`
	Path     string `json:"path"`     // HTTP path for health checks
	Port     int    `json:"port"`     // Port for health checks
	Protocol string `json:"protocol"` // "http", "https", "tcp"
	Interval int    `json:"interval"` // Seconds between health checks
	Timeout  int    `json:"timeout"`  // Timeout for health check
	Retries  int    `json:"retries"`  // Number of consecutive failures before unhealthy
}

// DeploymentPlan represents the planned AWS resources for deployment
type DeploymentPlan struct {
	TaskID    string                   `json:"task_id"`
	Resources []AWSResourceDefinition  `json:"resources"`
	EstimatedCost float64              `json:"estimated_cost"`
	Warnings      []string             `json:"warnings"`
	Prerequisites []string             `json:"prerequisites"`
}

// AWSResourceDefinition represents an AWS resource to be created/managed
type AWSResourceDefinition struct {
	Type       string                 `json:"type"`        // e.g., "AWS::ECS::Service"
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
	DependsOn  []string               `json:"depends_on"`
}

// TaskResult represents the outcome of processing a task
type TaskResult struct {
	TaskID     string                 `json:"task_id"`
	Status     string                 `json:"status"`     // "success", "failed", "needs_input"
	Message    string                 `json:"message"`
	DeployedResources []string        `json:"deployed_resources"`
	Warnings   []string               `json:"warnings"`
	NextSteps  []string               `json:"next_steps"`
	Metadata   map[string]interface{} `json:"metadata"`
}