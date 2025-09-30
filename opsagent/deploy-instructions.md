# BigFoot Golf OpsAgent - Deployment Instructions

## Prerequisites

1. **Install Pulumi CLI**
   ```bash
   curl -fsSL https://get.pulumi.com | sh
   ```

2. **Configure AWS Credentials**
   ```bash
   aws configure
   # OR set environment variables:
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_DEFAULT_REGION=us-east-1
   ```

3. **Install Go** (version 1.21 or later)

## Setup and Configuration

1. **Navigate to the opsagent directory**
   ```bash
   cd opsagent/
   ```

2. **Initialize Pulumi stack**
   ```bash
   pulumi stack init dev
   # OR
   pulumi stack init prod
   ```

3. **Configure required secrets** (use strong, unique passwords)
   ```bash
   # Neo4j password
   pulumi config set --secret neo4j-password "your-strong-neo4j-password"

   # Session key for web app authentication
   pulumi config set --secret session-key "your-32-character-session-key"

   # Gmail credentials for email functionality
   pulumi config set --secret gmail-pass "your-gmail-app-password"

   # Anthropic API key for AI features
   pulumi config set --secret anthropic-api-key "your-anthropic-api-key"
   ```

4. **Configure optional settings**
   ```bash
   # Set AWS region (default: us-east-1)
   pulumi config set aws:region us-east-1

   # Set application mode (default: prod)
   pulumi config set mode prod  # or "dev"

   # Set Gmail user (default: jrzesz@gmail.com)
   pulumi config set gmail-user "your-email@gmail.com"

   # Set custom Docker images (optional)
   pulumi config set neo4j-image "neo4j:5.15-community"
   pulumi config set web-app-image "your-registry/golf-web-app:latest"
   ```

## Deployment

1. **Preview the deployment**
   ```bash
   pulumi preview
   ```

2. **Deploy the infrastructure**
   ```bash
   pulumi up
   ```

3. **Wait for deployment to complete** (~10-15 minutes)

## Post-Deployment

### Accessing the Application

After deployment, Pulumi will output important information. To access your application:

1. **Get the public IP of your ECS task**
   ```bash
   # List running tasks
   aws ecs list-tasks --cluster bigfoot-golf

   # Get task details (replace TASK_ARN with actual ARN)
   aws ecs describe-tasks --cluster bigfoot-golf --tasks TASK_ARN
   ```

2. **Access the applications**
   - **Web Application**: `http://[PUBLIC_IP]:8000`
   - **Neo4j Browser**: `http://[PUBLIC_IP]:7474`

### Monitoring and Logs

1. **View ECS service status**
   ```bash
   aws ecs describe-services --cluster bigfoot-golf --services bigfoot-golf
   ```

2. **View CloudWatch logs**
   ```bash
   # Neo4j logs
   aws logs get-log-events --log-group-name "/ecs/bigfoot-golf-neo4j" --log-stream-name "ecs/neo4j/[TASK_ID]"

   # Web app logs
   aws logs get-log-events --log-group-name "/ecs/bigfoot-golf-web-app" --log-stream-name "ecs/web-app/[TASK_ID]"
   ```

### Scaling

To change the number of running instances:

```bash
aws ecs update-service --cluster bigfoot-golf --service bigfoot-golf --desired-count 2
```

## Configuration Details

### Environment Variables (automatically configured)

**Neo4j Container:**
- `NEO4J_PASSWORD`: Retrieved from AWS Secrets Manager

**Web App Container:**
- `DB_URI`: `bolt://localhost:7687`
- `DB_ADMIN`: Neo4j password from Secrets Manager
- `MODE`: Application mode (dev/prod)
- `GMAIL_USER`: Gmail username
- `GMAIL_PASS`: Gmail password from Secrets Manager
- `SESSION_KEY`: Session encryption key from Secrets Manager
- `ANTHROPIC_API_KEY`: Anthropic API key from Secrets Manager

### AWS Resources Created

- **ECS Cluster**: `bigfoot-golf`
- **ECS Service**: `bigfoot-golf`
- **EFS File System**: Persistent storage for Neo4j data
- **Security Groups**: Network access controls
- **IAM Roles**: Task execution and runtime permissions
- **Secrets Manager**: Secure credential storage
- **CloudWatch Log Groups**: Application logging

## Troubleshooting

### Common Issues

1. **Task fails to start**
   - Check CloudWatch logs for error messages
   - Verify all secrets are properly configured
   - Ensure Docker images are accessible

2. **Neo4j not accessible**
   - Wait for health checks to pass (can take 2-3 minutes)
   - Check security group rules
   - Verify EFS mount is successful

3. **Web app can't connect to Neo4j**
   - Ensure Neo4j container is healthy
   - Check internal networking (both containers in same task)
   - Verify database credentials

### Debugging Commands

```bash
# Check task status
aws ecs describe-tasks --cluster bigfoot-golf --tasks $(aws ecs list-tasks --cluster bigfoot-golf --query 'taskArns[0]' --output text)

# View task definition
aws ecs describe-task-definition --task-definition bigfoot-golf

# Check service events
aws ecs describe-services --cluster bigfoot-golf --services bigfoot-golf --query 'services[0].events'
```

## Cleanup

To destroy all resources:

```bash
pulumi destroy
```

**Warning**: This will permanently delete all data stored in the Neo4j database.

## Cost Optimization Notes

- **EFS**: Provisioned at 1 MiB/s throughput (minimal cost)
- **Fargate**: 1 vCPU, 2GB memory (can be adjusted in main.go)
- **No NAT Gateway**: Uses public subnets to reduce costs
- **CloudWatch**: 7-day log retention to control costs

## Security Features

- **EFS Encryption**: Transit and at-rest encryption enabled
- **Secrets Manager**: All sensitive data stored securely
- **IAM Roles**: Least-privilege access policies
- **Security Groups**: Restrictive network access
- **VPC**: Deployed in default VPC with security groups

For production deployments, consider:
- Using a custom VPC with private subnets
- Adding an Application Load Balancer for SSL termination
- Implementing CloudFront for caching and DDoS protection
- Setting up automated backups for the EFS volume