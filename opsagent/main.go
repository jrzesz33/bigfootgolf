package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/efs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/secretsmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get configuration values
		cfg := config.New(ctx, "")
		awsConfig := config.New(ctx, "aws")

		region := awsConfig.Get("region")
		if region == "" {
			region = "us-east-1"
		}

		neo4jPassword := cfg.RequireSecret("neo4j-password")
		sessionKey := cfg.RequireSecret("session-key")
		gmailUser := cfg.Get("gmail-user")
		if gmailUser == "" {
			gmailUser = "jrzesz@gmail.com"
		}
		gmailPass := cfg.RequireSecret("gmail-pass")
		anthropicAPIKey := cfg.RequireSecret("anthropic-api-key")
		neo4jImage := cfg.Get("neo4j-image")
		if neo4jImage == "" {
			neo4jImage = "neo4j:5.15-community"
		}
		webAppImage := cfg.Get("web-app-image")
		if webAppImage == "" {
			webAppImage = "your-registry/golf-web-app:latest"
		}
		mode := cfg.Get("mode")
		if mode == "" {
			mode = "prod"
		}

		// Get current AWS account info
		current, err := aws.GetCallerIdentity(ctx, nil, nil)
		if err != nil {
			return err
		}

		// Get default VPC and subnets
		defaultVpc, err := ec2.LookupVpc(ctx, &ec2.LookupVpcArgs{
			Default: pulumi.BoolRef(true),
		}, nil)
		if err != nil {
			return err
		}

		// Get subnets in the default VPC
		subnets, err := ec2.GetSubnets(ctx, &ec2.GetSubnetsArgs{
			Filters: []ec2.GetSubnetsFilter{
				{
					Name:   "vpc-id",
					Values: []string{defaultVpc.Id},
				},
			},
		}, nil)
		if err != nil {
			return err
		}

		// Create ECS Cluster
		cluster, err := ecs.NewCluster(ctx, "bigfoot-golf-cluster", &ecs.ClusterArgs{
			Name: pulumi.String("bigfoot-golf"),
			Settings: ecs.ClusterSettingArray{
				&ecs.ClusterSettingArgs{
					Name:  pulumi.String("containerInsights"),
					Value: pulumi.String("enabled"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("BigFoot Golf Cluster"),
				"Environment": pulumi.String(mode),
				"Project":     pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Create EFS File System for Neo4j persistent storage
		efsFileSystem, err := efs.NewFileSystem(ctx, "neo4j-efs", &efs.FileSystemArgs{
			CreationToken:                pulumi.String("bigfoot-golf-neo4j-data"),
			PerformanceMode:              pulumi.String("generalPurpose"),
			ThroughputMode:               pulumi.String("provisioned"),
			ProvisionedThroughputInMibps: pulumi.Float64(1),
			Encrypted:                    pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf Neo4j Data"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Create Security Group for EFS
		efsSecurityGroup, err := ec2.NewSecurityGroup(ctx, "efs-sg", &ec2.SecurityGroupArgs{
			Name:        pulumi.String("bigfoot-golf-efs-sg"),
			Description: pulumi.String("Security group for EFS access"),
			VpcId:       pulumi.String(defaultVpc.Id),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf EFS SG"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Create Security Group for ECS Tasks
		taskSecurityGroup, err := ec2.NewSecurityGroup(ctx, "task-sg", &ec2.SecurityGroupArgs{
			Name:        pulumi.String("bigfoot-golf-task-sg"),
			Description: pulumi.String("Security group for ECS tasks"),
			VpcId:       pulumi.String(defaultVpc.Id),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf Task SG"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Security Group Rules
		// Task accepts HTTP traffic on port 8000
		_, err = ec2.NewSecurityGroupRule(ctx, "task-web-ingress", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(8000),
			ToPort:          pulumi.Int(8000),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			SecurityGroupId: taskSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// Task accepts Neo4j HTTP traffic on port 7474
		_, err = ec2.NewSecurityGroupRule(ctx, "task-neo4j-http-ingress", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(7474),
			ToPort:          pulumi.Int(7474),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			SecurityGroupId: taskSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// Task internal communication on port 7687 (Neo4j Bolt)
		_, err = ec2.NewSecurityGroupRule(ctx, "task-neo4j-bolt-ingress", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(7687),
			ToPort:          pulumi.Int(7687),
			Protocol:        pulumi.String("tcp"),
			Self:            pulumi.Bool(true),
			SecurityGroupId: taskSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// EFS accepts NFS from tasks
		_, err = ec2.NewSecurityGroupRule(ctx, "efs-nfs-ingress", &ec2.SecurityGroupRuleArgs{
			Type:                  pulumi.String("ingress"),
			FromPort:              pulumi.Int(2049),
			ToPort:                pulumi.Int(2049),
			Protocol:              pulumi.String("tcp"),
			SourceSecurityGroupId: taskSecurityGroup.ID(),
			SecurityGroupId:       efsSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// Task allows all outbound traffic
		_, err = ec2.NewSecurityGroupRule(ctx, "task-all-egress", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("egress"),
			FromPort:        pulumi.Int(0),
			ToPort:          pulumi.Int(65535),
			Protocol:        pulumi.String("-1"),
			CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			SecurityGroupId: taskSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// Create Security Group for Application Load Balancer
		albSecurityGroup, err := ec2.NewSecurityGroup(ctx, "alb-sg", &ec2.SecurityGroupArgs{
			Name:        pulumi.String("bigfoot-golf-alb-sg"),
			Description: pulumi.String("Security group for Application Load Balancer"),
			VpcId:       pulumi.String(defaultVpc.Id),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf ALB SG"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// ALB Security Group Rules
		// Allow HTTP on port 80
		_, err = ec2.NewSecurityGroupRule(ctx, "alb-http-ingress", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(80),
			ToPort:          pulumi.Int(80),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			SecurityGroupId: albSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// Allow Neo4j HTTP on port 81
		_, err = ec2.NewSecurityGroupRule(ctx, "alb-neo4j-http-ingress", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(81),
			ToPort:          pulumi.Int(81),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			SecurityGroupId: albSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// Allow Neo4j Bolt on port 7687
		_, err = ec2.NewSecurityGroupRule(ctx, "alb-neo4j-bolt-ingress", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(7687),
			ToPort:          pulumi.Int(7687),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			SecurityGroupId: albSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// ALB egress to tasks
		_, err = ec2.NewSecurityGroupRule(ctx, "alb-all-egress", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("egress"),
			FromPort:        pulumi.Int(0),
			ToPort:          pulumi.Int(65535),
			Protocol:        pulumi.String("-1"),
			CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			SecurityGroupId: albSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// Update task security group to allow traffic from ALB
		_, err = ec2.NewSecurityGroupRule(ctx, "task-from-alb-ingress", &ec2.SecurityGroupRuleArgs{
			Type:                  pulumi.String("ingress"),
			FromPort:              pulumi.Int(0),
			ToPort:                pulumi.Int(65535),
			Protocol:              pulumi.String("tcp"),
			SourceSecurityGroupId: albSecurityGroup.ID(),
			SecurityGroupId:       taskSecurityGroup.ID(),
		})
		if err != nil {
			return err
		}

		// Create Application Load Balancer
		alb, err := lb.NewLoadBalancer(ctx, "bigfoot-golf-alb", &lb.LoadBalancerArgs{
			Name:             pulumi.String("bigfoot-golf-alb"),
			Internal:         pulumi.Bool(false),
			LoadBalancerType: pulumi.String("application"),
			SecurityGroups:   pulumi.StringArray{albSecurityGroup.ID()},
			Subnets:          pulumi.ToStringArray(subnets.Ids),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf ALB"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Create Target Group for Web App (port 8000)
		webAppTargetGroup, err := lb.NewTargetGroup(ctx, "web-app-tg", &lb.TargetGroupArgs{
			Name:       pulumi.String("bigfoot-golf-web-app-tg"),
			Port:       pulumi.Int(8000),
			Protocol:   pulumi.String("HTTP"),
			VpcId:      pulumi.String(defaultVpc.Id),
			TargetType: pulumi.String("ip"),
			HealthCheck: &lb.TargetGroupHealthCheckArgs{
				Enabled:            pulumi.Bool(true),
				Path:               pulumi.String("/web/app.css"),
				Port:               pulumi.String("8000"),
				Protocol:           pulumi.String("HTTP"),
				HealthyThreshold:   pulumi.Int(2),
				UnhealthyThreshold: pulumi.Int(3),
				Timeout:            pulumi.Int(5),
				Interval:           pulumi.Int(30),
				Matcher:            pulumi.String("200-299"),
			},
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf Web App TG"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Create Target Group for Neo4j HTTP (port 7474)
		neo4jHttpTargetGroup, err := lb.NewTargetGroup(ctx, "neo4j-http-tg", &lb.TargetGroupArgs{
			Name:       pulumi.String("bigfoot-golf-neo4j-http-tg"),
			Port:       pulumi.Int(7474),
			Protocol:   pulumi.String("HTTP"),
			VpcId:      pulumi.String(defaultVpc.Id),
			TargetType: pulumi.String("ip"),
			HealthCheck: &lb.TargetGroupHealthCheckArgs{
				Enabled:            pulumi.Bool(true),
				Path:               pulumi.String("/"),
				Port:               pulumi.String("7474"),
				Protocol:           pulumi.String("HTTP"),
				HealthyThreshold:   pulumi.Int(2),
				UnhealthyThreshold: pulumi.Int(3),
				Timeout:            pulumi.Int(5),
				Interval:           pulumi.Int(30),
				Matcher:            pulumi.String("200-299"),
			},
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf Neo4j HTTP TG"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Create Listener for port 80 -> Web App (8000)
		_, err = lb.NewListener(ctx, "web-app-listener", &lb.ListenerArgs{
			LoadBalancerArn: alb.Arn,
			Port:            pulumi.Int(80),
			Protocol:        pulumi.String("HTTP"),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: webAppTargetGroup.Arn,
				},
			},
		})
		if err != nil {
			return err
		}

		// Create Listener for port 81 -> Neo4j HTTP (7474)
		_, err = lb.NewListener(ctx, "neo4j-http-listener", &lb.ListenerArgs{
			LoadBalancerArn: alb.Arn,
			Port:            pulumi.Int(81),
			Protocol:        pulumi.String("HTTP"),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: neo4jHttpTargetGroup.Arn,
				},
			},
		})
		if err != nil {
			return err
		}

		// Create EFS Mount Targets and collect them
		var mountTargets []*efs.MountTarget
		for i, subnetId := range subnets.Ids {
			mountTarget, err := efs.NewMountTarget(ctx, fmt.Sprintf("neo4j-efs-mount-%d", i), &efs.MountTargetArgs{
				FileSystemId:   efsFileSystem.ID(),
				SubnetId:       pulumi.String(subnetId),
				SecurityGroups: pulumi.StringArray{efsSecurityGroup.ID()},
			})
			if err != nil {
				return err
			}
			mountTargets = append(mountTargets, mountTarget)
		}

		// Helper function to get or create a secret
		getOrCreateSecret := func(resourceName, secretName, description string) (*secretsmanager.Secret, error) {
			// Try to look up existing secret
			existingSecret, lookupErr := secretsmanager.LookupSecret(ctx, &secretsmanager.LookupSecretArgs{
				Name: pulumi.StringRef(secretName),
			}, nil)

			// If secret exists, import it as a Pulumi resource
			if lookupErr == nil && existingSecret != nil {
				return secretsmanager.GetSecret(ctx, resourceName, pulumi.ID(existingSecret.Arn), &secretsmanager.SecretState{
					Name:        pulumi.String(secretName),
					Description: pulumi.String(description),
				}, pulumi.Protect(true))
			}

			// Secret doesn't exist, create it with protection
			// Set RecoveryWindowInDays to 0 to force immediate deletion if needed
			// Also set ForceOverwriteReplicaSecret to handle edge cases
			return secretsmanager.NewSecret(ctx, resourceName, &secretsmanager.SecretArgs{
				Name:                        pulumi.String(secretName),
				Description:                 pulumi.String(description),
				RecoveryWindowInDays:        pulumi.IntPtr(7), // 7-day recovery window (minimum for protection)
				ForceOverwriteReplicaSecret: pulumi.BoolPtr(true),
				Tags: pulumi.StringMap{
					"Name":    pulumi.String(description),
					"Project": pulumi.String("BigFoot Golf"),
				},
			}, pulumi.Protect(true))
		}

		// Create or get Secrets Manager secrets
		neo4jSecret, err := getOrCreateSecret("neo4j-password-secret", "bigfoot-golf/neo4j-password", "Neo4j database password")
		if err != nil {
			return err
		}

		// Only create/update the version if the secret was newly created or needs updating
		_, err = secretsmanager.NewSecretVersion(ctx, "neo4j-password-version", &secretsmanager.SecretVersionArgs{
			SecretId:     neo4jSecret.ID(),
			SecretString: neo4jPassword,
		}, pulumi.IgnoreChanges([]string{"secretString"}))
		if err != nil {
			return err
		}

		sessionSecret, err := getOrCreateSecret("session-key-secret", "bigfoot-golf/session-key", "Session key for web application")
		if err != nil {
			return err
		}

		_, err = secretsmanager.NewSecretVersion(ctx, "session-key-version", &secretsmanager.SecretVersionArgs{
			SecretId:     sessionSecret.ID(),
			SecretString: sessionKey,
		}, pulumi.IgnoreChanges([]string{"secretString"}))
		if err != nil {
			return err
		}

		gmailPassSecret, err := getOrCreateSecret("gmail-pass-secret", "bigfoot-golf/gmail-pass", "Gmail password for email access")
		if err != nil {
			return err
		}

		_, err = secretsmanager.NewSecretVersion(ctx, "gmail-pass-version", &secretsmanager.SecretVersionArgs{
			SecretId:     gmailPassSecret.ID(),
			SecretString: gmailPass,
		}, pulumi.IgnoreChanges([]string{"secretString"}))
		if err != nil {
			return err
		}

		anthropicSecret, err := getOrCreateSecret("anthropic-api-key-secret", "bigfoot-golf/anthropic-api-key", "Anthropic API key")
		if err != nil {
			return err
		}

		_, err = secretsmanager.NewSecretVersion(ctx, "anthropic-api-key-version", &secretsmanager.SecretVersionArgs{
			SecretId:     anthropicSecret.ID(),
			SecretString: anthropicAPIKey,
		}, pulumi.IgnoreChanges([]string{"secretString"}))
		if err != nil {
			return err
		}

		// Create IAM Role for ECS Task Execution
		taskExecutionRole, err := iam.NewRole(ctx, "task-execution-role", &iam.RoleArgs{
			Name: pulumi.String("bigfoot-golf-task-execution-role"),
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Action": "sts:AssumeRole",
						"Effect": "Allow",
						"Principal": {
							"Service": "ecs-tasks.amazonaws.com"
						}
					}
				]
			}`),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf Task Execution Role"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Attach managed policy for ECS task execution
		_, err = iam.NewRolePolicyAttachment(ctx, "task-execution-role-policy", &iam.RolePolicyAttachmentArgs{
			Role:      taskExecutionRole.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
		})
		if err != nil {
			return err
		}

		// Create custom policy for secrets access
		secretsPolicy, err := iam.NewPolicy(ctx, "secrets-policy", &iam.PolicyArgs{
			Name:        pulumi.String("bigfoot-golf-secrets-policy"),
			Description: pulumi.String("Policy for accessing BigFoot Golf secrets"),
			Policy: pulumi.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Action": [
							"secretsmanager:GetSecretValue"
						],
						"Resource": [
							"%s",
							"%s",
							"%s",
							"%s"
						]
					}
				]
			}`, neo4jSecret.Arn, sessionSecret.Arn, gmailPassSecret.Arn, anthropicSecret.Arn),
		})
		if err != nil {
			return err
		}

		// Attach secrets policy to execution role
		_, err = iam.NewRolePolicyAttachment(ctx, "task-execution-secrets-policy", &iam.RolePolicyAttachmentArgs{
			Role:      taskExecutionRole.Name,
			PolicyArn: secretsPolicy.Arn,
		})
		if err != nil {
			return err
		}

		// Create custom policy for CloudWatch Logs access
		logsPolicy, err := iam.NewPolicy(ctx, "logs-policy", &iam.PolicyArgs{
			Name:        pulumi.String("bigfoot-golf-logs-policy"),
			Description: pulumi.String("Policy for CloudWatch Logs access"),
			Policy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Action": [
							"logs:CreateLogGroup",
							"logs:CreateLogStream",
							"logs:PutLogEvents"
						],
						"Resource": "arn:aws:logs:*:*:*"
					}
				]
			}`),
		})
		if err != nil {
			return err
		}

		// Attach logs policy to execution role
		_, err = iam.NewRolePolicyAttachment(ctx, "task-execution-logs-policy", &iam.RolePolicyAttachmentArgs{
			Role:      taskExecutionRole.Name,
			PolicyArn: logsPolicy.Arn,
		})
		if err != nil {
			return err
		}

		// Create IAM Role for ECS Task
		taskRole, err := iam.NewRole(ctx, "task-role", &iam.RoleArgs{
			Name: pulumi.String("bigfoot-golf-task-role"),
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Action": "sts:AssumeRole",
						"Effect": "Allow",
						"Principal": {
							"Service": "ecs-tasks.amazonaws.com"
						}
					}
				]
			}`),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf Task Role"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Create EFS access policy for task role
		efsPolicy, err := iam.NewPolicy(ctx, "efs-policy", &iam.PolicyArgs{
			Name:        pulumi.String("bigfoot-golf-efs-policy"),
			Description: pulumi.String("Policy for EFS access"),
			Policy: pulumi.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Action": [
							"elasticfilesystem:ClientMount",
							"elasticfilesystem:ClientWrite",
							"elasticfilesystem:ClientRootAccess"
						],
						"Resource": "%s"
					}
				]
			}`, efsFileSystem.Arn),
		})
		if err != nil {
			return err
		}

		// Attach EFS policy to task role
		_, err = iam.NewRolePolicyAttachment(ctx, "task-efs-policy", &iam.RolePolicyAttachmentArgs{
			Role:      taskRole.Name,
			PolicyArn: efsPolicy.Arn,
		})
		if err != nil {
			return err
		}

		// Create ECS Task Definition - only container definitions as JSON
		containerDefinitionsJSON := pulumi.All(
			neo4jSecret.Arn,
			sessionSecret.Arn,
			gmailPassSecret.Arn,
			anthropicSecret.Arn,
		).ApplyT(func(args []interface{}) (string, error) {
			neo4jSecretArn := args[0].(string)
			sessionSecretArn := args[1].(string)
			gmailPassSecretArn := args[2].(string)
			anthropicSecretArn := args[3].(string)

			containerDefs := []map[string]interface{}{
				{
					"name":      "neo4j",
					"image":     neo4jImage,
					"essential": true,
					"portMappings": []map[string]interface{}{
						{
							"containerPort": 7474,
							"protocol":      "tcp",
						},
						{
							"containerPort": 7687,
							"protocol":      "tcp",
						},
					},
					"environment": []map[string]interface{}{
						{
							"name":  "NEO4J_AUTH",
							"value": "neo4j/TEMP_PASSWORD_TO_BE_REPLACED",
						},
					},
					"secrets": []map[string]interface{}{
						{
							"name":      "NEO4J_PASSWORD",
							"valueFrom": neo4jSecretArn,
						},
					},
					"mountPoints": []map[string]interface{}{
						{
							"sourceVolume":  "neo4j-data-volume",
							"containerPath": "/data",
							"readOnly":      false,
						},
					},
					"logConfiguration": map[string]interface{}{
						"logDriver": "awslogs",
						"options": map[string]interface{}{
							"awslogs-group":         "/ecs/bigfoot-golf-neo4j",
							"awslogs-region":        region,
							"awslogs-stream-prefix": "ecs",
							"awslogs-create-group":  "true",
						},
					},
					"healthCheck": map[string]interface{}{
						"command": []string{
							"CMD-SHELL",
							"wget --no-verbose --tries=1 --spider http://localhost:7474/ || exit 1",
						},
						"interval":    30,
						"timeout":     3,
						"retries":     3,
						"startPeriod": 60,
					},
				},
				{
					"name":      "web-app",
					"image":     webAppImage,
					"essential": true,
					"portMappings": []map[string]interface{}{
						{
							"containerPort": 8000,
							"protocol":      "tcp",
						},
					},
					"environment": []map[string]interface{}{
						{
							"name":  "DB_URI",
							"value": "bolt://localhost:7687",
						},
						{
							"name":  "MODE",
							"value": mode,
						},
						{
							"name":  "GMAIL_USER",
							"value": gmailUser,
						},
					},
					"secrets": []map[string]interface{}{
						{
							"name":      "DB_ADMIN",
							"valueFrom": neo4jSecretArn,
						},
						{
							"name":      "SESSION_KEY",
							"valueFrom": sessionSecretArn,
						},
						{
							"name":      "GMAIL_PASS",
							"valueFrom": gmailPassSecretArn,
						},
						{
							"name":      "ANTHROPIC_API_KEY",
							"valueFrom": anthropicSecretArn,
						},
					},
					"dependsOn": []map[string]interface{}{
						{
							"containerName": "neo4j",
							"condition":     "HEALTHY",
						},
					},
					"logConfiguration": map[string]interface{}{
						"logDriver": "awslogs",
						"options": map[string]interface{}{
							"awslogs-group":         "/ecs/bigfoot-golf-web-app",
							"awslogs-region":        region,
							"awslogs-stream-prefix": "ecs",
							"awslogs-create-group":  "true",
						},
					},
					"healthCheck": map[string]interface{}{
						"command": []string{
							"CMD-SHELL",
							"wget --no-verbose --tries=1 --spider http://localhost:8000/web/app.css || exit 1",
						},
						"interval":    30,
						"timeout":     5,
						"retries":     3,
						"startPeriod": 60,
					},
				},
			}

			jsonBytes, err := json.Marshal(containerDefs)
			if err != nil {
				return "", err
			}
			return string(jsonBytes), nil
		}).(pulumi.StringOutput)

		// Build dependencies on all mount targets
		var mountTargetDeps []pulumi.Resource
		for _, mt := range mountTargets {
			mountTargetDeps = append(mountTargetDeps, mt)
		}

		taskDefinition, err := ecs.NewTaskDefinition(ctx, "bigfoot-golf-task", &ecs.TaskDefinitionArgs{
			Family:                  pulumi.String("bigfoot-golf"),
			ContainerDefinitions:    containerDefinitionsJSON,
			RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
			NetworkMode:             pulumi.String("awsvpc"),
			Cpu:                     pulumi.String("1024"),
			Memory:                  pulumi.String("2048"),
			ExecutionRoleArn:        taskExecutionRole.Arn,
			TaskRoleArn:             taskRole.Arn,
			Volumes: ecs.TaskDefinitionVolumeArray{
				&ecs.TaskDefinitionVolumeArgs{
					Name: pulumi.String("neo4j-data-volume"),
					EfsVolumeConfiguration: &ecs.TaskDefinitionVolumeEfsVolumeConfigurationArgs{
						FileSystemId:      efsFileSystem.ID(),
						TransitEncryption: pulumi.String("ENABLED"),
						AuthorizationConfig: &ecs.TaskDefinitionVolumeEfsVolumeConfigurationAuthorizationConfigArgs{
							Iam: pulumi.String("ENABLED"),
						},
					},
				},
			},
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf Task Definition"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		}, pulumi.DependsOn(mountTargetDeps))
		if err != nil {
			return err
		}

		// Create ECS Service with load balancer integration
		service, err := ecs.NewService(ctx, "bigfoot-golf-service", &ecs.ServiceArgs{
			Name:            pulumi.String("bigfoot-golf"),
			Cluster:         cluster.ID(),
			TaskDefinition:  taskDefinition.Arn,
			LaunchType:      pulumi.String("FARGATE"),
			DesiredCount:    pulumi.Int(1),
			PlatformVersion: pulumi.String("LATEST"),
			NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
				Subnets:        pulumi.ToStringArray(subnets.Ids),
				SecurityGroups: pulumi.StringArray{taskSecurityGroup.ID()},
				AssignPublicIp: pulumi.Bool(true),
			},
			LoadBalancers: ecs.ServiceLoadBalancerArray{
				&ecs.ServiceLoadBalancerArgs{
					TargetGroupArn: webAppTargetGroup.Arn,
					ContainerName:  pulumi.String("web-app"),
					ContainerPort:  pulumi.Int(8000),
				},
				&ecs.ServiceLoadBalancerArgs{
					TargetGroupArn: neo4jHttpTargetGroup.Arn,
					ContainerName:  pulumi.String("neo4j"),
					ContainerPort:  pulumi.Int(7474),
				},
			},
			Tags: pulumi.StringMap{
				"Name":    pulumi.String("BigFoot Golf Service"),
				"Project": pulumi.String("BigFoot Golf"),
			},
		})
		if err != nil {
			return err
		}

		// Export outputs
		ctx.Export("clusterName", cluster.Name)
		ctx.Export("clusterArn", cluster.Arn)
		ctx.Export("serviceName", service.Name)
		//ctx.Export("serviceArn", service.Arn)
		ctx.Export("taskDefinitionArn", taskDefinition.Arn)
		ctx.Export("efsFileSystemId", efsFileSystem.ID())
		ctx.Export("efsFileSystemArn", efsFileSystem.Arn)

		// Export secrets ARNs for reference
		ctx.Export("neo4jSecretArn", neo4jSecret.Arn)
		ctx.Export("sessionSecretArn", sessionSecret.Arn)
		ctx.Export("gmailSecretArn", gmailPassSecret.Arn)
		ctx.Export("anthropicSecretArn", anthropicSecret.Arn)

		// Export configuration details
		ctx.Export("region", pulumi.String(region))
		ctx.Export("accountId", pulumi.String(current.AccountId))
		ctx.Export("mode", pulumi.String(mode))

		// Export load balancer endpoints
		ctx.Export("albDnsName", alb.DnsName)
		ctx.Export("webAppUrl", pulumi.Sprintf("http://%s", alb.DnsName))
		ctx.Export("neo4jHttpUrl", pulumi.Sprintf("http://%s:81", alb.DnsName))

		return nil
	})
}
