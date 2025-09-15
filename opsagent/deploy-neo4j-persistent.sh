#!/bin/bash

# Deploy Neo4j with persistent storage and public endpoint
# This script creates all necessary AWS resources for a production-ready Neo4j deployment

set -e

# Configuration
AWS_REGION="us-east-1"
PROJECT_NAME="bigfootgolf-opsagent"
ECS_CLUSTER_NAME="bigfootgolf-opsagent-cluster"

echo "🚀 Starting Neo4j deployment with persistent storage..."

# Step 1: Create EFS file system for persistent data
echo "📁 Creating EFS file system for Neo4j data..."
EFS_ID=$(aws efs create-file-system \
    --creation-token "bigfootgolf-neo4j-data-$(date +%s)" \
    --performance-mode generalPurpose \
    --throughput-mode provisioned \
    --provisioned-throughput-in-mibps 1 \
    --encrypted \
    --tags "Key=Project,Value=$PROJECT_NAME" "Key=Component,Value=neo4j-storage" \
    --query 'FileSystemId' \
    --output text)

echo "✅ EFS created: $EFS_ID"

# Step 2: Create EFS mount targets in each subnet
echo "🔗 Creating EFS mount targets..."
SUBNETS=("subnet-044f8a2be0880bb6f" "subnet-040ef165d121d0577" "subnet-01fe36a31868d14d6")
for subnet in "${SUBNETS[@]}"; do
    aws efs create-mount-target \
        --file-system-id $EFS_ID \
        --subnet-id $subnet \
        --security-groups sg-XXXXXXXXX
done

# Step 3: Create security groups
echo "🔒 Creating security groups..."

# ALB Security Group
ALB_SG_ID=$(aws ec2 create-security-group \
    --group-name "bigfootgolf-neo4j-alb-sg" \
    --description "Security group for Neo4j Application Load Balancer" \
    --vpc-id "vpc-0900c8f70d2ce2a18" \
    --tag-specifications "ResourceType=security-group,Tags=[{Key=Project,Value=$PROJECT_NAME},{Key=Component,Value=neo4j-alb-security-group}]" \
    --query 'GroupId' \
    --output text)

# Task Security Group  
TASK_SG_ID=$(aws ec2 create-security-group \
    --group-name "bigfootgolf-neo4j-task-sg" \
    --description "Security group for Neo4j ECS tasks" \
    --vpc-id "vpc-0900c8f70d2ce2a18" \
    --tag-specifications "ResourceType=security-group,Tags=[{Key=Project,Value=$PROJECT_NAME},{Key=Component,Value=neo4j-task-security-group}]" \
    --query 'GroupId' \
    --output text)

# Configure ALB security group rules
aws ec2 authorize-security-group-ingress \
    --group-id $ALB_SG_ID \
    --protocol tcp \
    --port 80 \
    --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
    --group-id $ALB_SG_ID \
    --protocol tcp \
    --port 443 \
    --cidr 0.0.0.0/0

# Configure task security group rules
aws ec2 authorize-security-group-ingress \
    --group-id $TASK_SG_ID \
    --protocol tcp \
    --port 7474 \
    --source-group $ALB_SG_ID

aws ec2 authorize-security-group-ingress \
    --group-id $TASK_SG_ID \
    --protocol tcp \
    --port 7687 \
    --source-group $ALB_SG_ID

echo "✅ Security groups created: ALB=$ALB_SG_ID, Task=$TASK_SG_ID"

# Step 4: Update task definition with EFS file system ID
echo "📝 Updating task definition..."
sed "s/fs-XXXXXXXXX/$EFS_ID/g" tasks/neodb-with-persistence.json > /tmp/task-definition.json

# Register task definition
TASK_DEF_ARN=$(aws ecs register-task-definition \
    --cli-input-json file:///tmp/task-definition.json \
    --query 'taskDefinition.taskDefinitionArn' \
    --output text)

echo "✅ Task definition registered: $TASK_DEF_ARN"

# Step 5: Create Application Load Balancer
echo "🌐 Creating Application Load Balancer..."
ALB_ARN=$(aws elbv2 create-load-balancer \
    --name "bigfootgolf-neo4j-alb" \
    --subnets "${SUBNETS[@]}" \
    --security-groups $ALB_SG_ID \
    --scheme internet-facing \
    --type application \
    --ip-address-type ipv4 \
    --tags "Key=Project,Value=$PROJECT_NAME" "Key=Component,Value=neo4j-load-balancer" \
    --query 'LoadBalancers[0].LoadBalancerArn' \
    --output text)

# Get ALB DNS name
ALB_DNS=$(aws elbv2 describe-load-balancers \
    --load-balancer-arns $ALB_ARN \
    --query 'LoadBalancers[0].DNSName' \
    --output text)

echo "✅ ALB created: $ALB_DNS"

# Step 6: Create Target Group
echo "🎯 Creating target group..."
TARGET_GROUP_ARN=$(aws elbv2 create-target-group \
    --name "bigfootgolf-neo4j-tg" \
    --protocol HTTP \
    --port 7474 \
    --vpc-id "vpc-0900c8f70d2ce2a18" \
    --target-type ip \
    --health-check-protocol HTTP \
    --health-check-path "/" \
    --health-check-port 7474 \
    --health-check-interval-seconds 30 \
    --health-check-timeout-seconds 5 \
    --healthy-threshold-count 2 \
    --unhealthy-threshold-count 3 \
    --tags "Key=Project,Value=$PROJECT_NAME" "Key=Component,Value=neo4j-target-group" \
    --query 'TargetGroups[0].TargetGroupArn' \
    --output text)

echo "✅ Target group created: $TARGET_GROUP_ARN"

# Step 7: Create ALB Listener
echo "👂 Creating ALB listener..."
aws elbv2 create-listener \
    --load-balancer-arn $ALB_ARN \
    --protocol HTTP \
    --port 80 \
    --default-actions "Type=forward,TargetGroupArn=$TARGET_GROUP_ARN"

echo "✅ Listener created"

# Step 8: Create ECS Service
echo "🚀 Creating ECS service..."
# Update service configuration with actual values
sed -e "s/REPLACE_WITH_TASK_SG_ID/$TASK_SG_ID/g" \
    -e "s/REPLACE_WITH_TARGET_GROUP_ARN/$TARGET_GROUP_ARN/g" \
    infrastructure/ecs-service.json > /tmp/ecs-service.json

aws ecs create-service \
    --cli-input-json file:///tmp/ecs-service.json

echo "✅ ECS service created"

# Step 9: Wait for deployment to complete
echo "⏳ Waiting for service to become stable..."
aws ecs wait services-stable \
    --cluster $ECS_CLUSTER_NAME \
    --services "bigfootgolf-neo4j-service"

echo "🎉 Deployment complete!"
echo ""
echo "📍 Access your Neo4j browser at: http://$ALB_DNS"
echo "🔐 Database password is stored in AWS Secrets Manager"
echo "💾 Data is persisted in EFS: $EFS_ID"
echo ""
echo "🔧 Resource Summary:"
echo "  - EFS File System: $EFS_ID"
echo "  - Load Balancer: $ALB_DNS"
echo "  - Task Definition: $TASK_DEF_ARN"
echo "  - Target Group: $TARGET_GROUP_ARN"
echo "  - ALB Security Group: $ALB_SG_ID"
echo "  - Task Security Group: $TASK_SG_ID"