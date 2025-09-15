# BigFoot Golf OpsAgent - Neo4j Deployment

This directory contains configuration files and deployment scripts for deploying Neo4j with persistent storage and a public endpoint on AWS ECS Fargate.

## 🏗️ **Architecture Overview**

```
Internet → ALB (Port 80) → ECS Tasks (Port 7474) → Neo4j Container
                                      ↓
                                 EFS Volume (/data)
```

### **Components:**
- **ECS Fargate**: Serverless container hosting
- **Application Load Balancer**: Public endpoint with health checks
- **EFS**: Persistent storage for Neo4j `/data` directory
- **Secrets Manager**: Secure password storage
- **Security Groups**: Network security controls

## 📁 **File Structure**

```
opsagent/
├── tasks/
│   ├── neodb.json                    # Original task definition
│   └── neodb-with-persistence.json   # Enhanced with EFS volume
├── infrastructure/
│   ├── alb-config.json              # Load balancer configuration
│   ├── security-group.json          # Security group definitions
│   └── ecs-service.json             # ECS service with ALB integration
├── deploy-neo4j-persistent.sh       # Complete deployment script
└── README.md                        # This file
```

## 🚀 **Quick Deployment**

### **Prerequisites:**
- AWS CLI configured with appropriate permissions
- ECS cluster `bigfootgolf-opsagent-cluster` exists
- IAM role `ecsTaskExecutionRole` exists
- Secrets Manager secret for Neo4j password exists

### **Deploy:**
```bash
cd /workspaces/golf_app/opsagent
./deploy-neo4j-persistent.sh
```

This script will:
1. ✅ Create EFS file system with encryption
2. ✅ Set up EFS mount targets in multiple AZs
3. ✅ Create security groups for ALB and tasks
4. ✅ Register enhanced task definition with EFS volume
5. ✅ Create Application Load Balancer
6. ✅ Set up target group and health checks
7. ✅ Create ECS service with ALB integration
8. ✅ Wait for deployment to complete

## 🔧 **Key Modifications Made**

### **1. Persistent Storage (EFS Volume)**
```json
{
  "volumes": [
    {
      "name": "neo4j-data-volume",
      "efsVolumeConfiguration": {
        "fileSystemId": "fs-XXXXXXXXX",
        "transitEncryption": "ENABLED",
        "authorizationConfig": {
          "iam": "ENABLED"
        }
      }
    }
  ]
}
```

### **2. Container Mount Points**
```json
{
  "mountPoints": [
    {
      "sourceVolume": "neo4j-data-volume",
      "containerPath": "/data",
      "readOnly": false
    }
  ]
}
```

### **3. Load Balancer Integration**
- **Public endpoint**: `http://<ALB-DNS-NAME>`
- **Health checks**: HTTP on port 7474
- **Target type**: IP (required for Fargate)
- **Multi-AZ**: Deployed across 3 availability zones

### **4. Security Configuration**
- **ALB Security Group**: Allows HTTP/HTTPS from internet
- **Task Security Group**: Allows access only from ALB
- **EFS Security**: Transit encryption enabled

## 🌐 **Public Access**

After deployment completes:

1. **Neo4j Browser**: `http://<ALB-DNS-NAME>`
2. **Database Connection**: Use ALB DNS name on port 80
3. **Password**: Retrieved from AWS Secrets Manager
4. **Data Persistence**: All data stored in EFS survives container restarts

## 📊 **Monitoring & Logs**

- **CloudWatch Logs**: `/ecs/bigfootgolf-task-v3-database`
- **ALB Metrics**: Available in CloudWatch
- **Health Checks**: Configured for port 7474
- **EFS Metrics**: File system performance monitoring

## 🔒 **Security Features**

- ✅ EFS encryption at rest and in transit
- ✅ Secrets Manager for password storage
- ✅ Security groups with least privilege access
- ✅ ALB with AWS security features
- ✅ Private subnets for ECS tasks (with NAT gateway)

## 💰 **Cost Optimization**

- **EFS**: Provisioned throughput at 1 MiB/s (minimal cost)
- **Fargate**: 512 CPU units, 1024 MB memory
- **ALB**: Standard Application Load Balancer pricing
- **No NAT Gateway**: Uses public subnets with public IPs

## 🔧 **Manual Customization**

If you need to modify the configuration:

1. **EFS File System ID**: Update in `neodb-with-persistence.json`
2. **Security Group IDs**: Update in deployment script
3. **Subnet IDs**: Already configured for your default VPC
4. **Resource Names**: Modify in deployment script

## 🎯 **Next Steps**

After deployment:
1. Access Neo4j browser at the provided ALB DNS name
2. Login with credentials from Secrets Manager
3. Your data will persist across container restarts
4. Scale the service up/down as needed

The deployment provides a production-ready Neo4j instance with persistent storage and public access through a secure, load-balanced endpoint.