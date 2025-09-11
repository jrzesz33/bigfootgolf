package aws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/sirupsen/logrus"
)

// Manager handles AWS service interactions
type Manager struct {
	region         string
	cloudControl   *cloudcontrol.Client
	ec2            *ec2.Client
	ecs            *ecs.Client
}

// NewManager creates a new AWS manager
func NewManager(region string) (*Manager, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Manager{
		region:       region,
		cloudControl: cloudcontrol.NewFromConfig(cfg),
		ec2:          ec2.NewFromConfig(cfg),
		ecs:          ecs.NewFromConfig(cfg),
	}, nil
}

// CreateResource creates an AWS resource using Cloud Control API
func (m *Manager) CreateResource(ctx context.Context, resourceType string, properties map[string]interface{}) (string, error) {
	logrus.WithFields(logrus.Fields{
		"resource_type": resourceType,
		"region":        m.region,
	}).Info("Creating AWS resource")

	// Convert properties to JSON string for Cloud Control API
	propertiesJSON, err := marshalProperties(properties)
	if err != nil {
		return "", fmt.Errorf("failed to marshal properties: %w", err)
	}

	input := &cloudcontrol.CreateResourceInput{
		TypeName:              &resourceType,
		DesiredState:          &propertiesJSON,
		ClientToken:           generateClientToken(),
	}

	resp, err := m.cloudControl.CreateResource(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create resource %s: %w", resourceType, err)
	}

	logrus.WithFields(logrus.Fields{
		"resource_type": resourceType,
		"resource_id":   *resp.ProgressEvent.Identifier,
	}).Info("AWS resource creation initiated")

	return *resp.ProgressEvent.Identifier, nil
}

// GetResourceStatus checks the status of a resource creation/update operation
func (m *Manager) GetResourceStatus(ctx context.Context, requestToken string) (string, error) {
	input := &cloudcontrol.GetResourceRequestStatusInput{
		RequestToken: &requestToken,
	}

	resp, err := m.cloudControl.GetResourceRequestStatus(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get resource status: %w", err)
	}

	return string(resp.ProgressEvent.OperationStatus), nil
}

// ListECSClusters lists available ECS clusters (for cost optimization)
func (m *Manager) ListECSClusters(ctx context.Context) ([]string, error) {
	resp, err := m.ecs.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ECS clusters: %w", err)
	}

	clusters := make([]string, len(resp.ClusterArns))
	for i, arn := range resp.ClusterArns {
		clusters[i] = arn
	}

	return clusters, nil
}

// GetVPCInfo retrieves VPC information for networking setup
func (m *Manager) GetVPCInfo(ctx context.Context) (*VPCInfo, error) {
	// Get default VPC to minimize costs
	resp, err := m.ec2.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("is-default"),
				Values: []string{"true"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	if len(resp.Vpcs) == 0 {
		return nil, fmt.Errorf("no default VPC found")
	}

	vpc := resp.Vpcs[0]
	
	// Get subnets for the VPC
	subnetsResp, err := m.ec2.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{*vpc.VpcId},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}

	subnets := make([]string, len(subnetsResp.Subnets))
	for i, subnet := range subnetsResp.Subnets {
		subnets[i] = *subnet.SubnetId
	}

	return &VPCInfo{
		VpcID:   *vpc.VpcId,
		Subnets: subnets,
	}, nil
}

// VPCInfo contains VPC networking information
type VPCInfo struct {
	VpcID   string   `json:"vpc_id"`
	Subnets []string `json:"subnets"`
}

// Helper functions

func marshalProperties(properties map[string]interface{}) (string, error) {
	if properties == nil {
		return "{}", nil
	}
	
	data, err := json.Marshal(properties)
	if err != nil {
		return "", fmt.Errorf("failed to marshal properties to JSON: %w", err)
	}
	
	return string(data), nil
}

func generateClientToken() *string {
	// Generate a unique client token for idempotency
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based token
		token := fmt.Sprintf("boat-client-%d", time.Now().UnixNano())
		return &token
	}
	
	token := "boat-" + hex.EncodeToString(bytes)
	return &token
}

