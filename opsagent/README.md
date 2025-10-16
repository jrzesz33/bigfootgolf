# BigFoot Golf OpsAgent - Neo4j Deployment

This directory contains a Pulumi Application that is responsible for operating the infrastructure and deployment of the Golf Booking application which includes a  Neo4j database with persistent storage and a public endpoint on AWS ECS Fargate.

## 🏗️ **Architecture Overview**
- **DB Application**: neo4j database container  
  - EFS Storage
  - Web Application Access Connections via Ports 7474 and 7687
- **Web Application**: web application container
  - Web Application Container Serves Web Traffic from Port 8000
- **Load Balancer**: Public Access to Web Application
### **Components:**
- **ECS Fargate**: Serverless container hosting
- **Application Load Balancer**: Public endpoint with health checks
- **EFS**: Persistent storage for Neo4j `/data` directory
- **Secrets Manager**: Secure password storage
- **Security Groups**: Network security controls

## 🚀 **Quick Deployment**

### **Prerequisites:**
- Pulumi with pulumi-aws and the go SDK to Manage Infrastructure
- Ability to access the AWS via AWS Credentials
- Required Environment Variables and Secrets for Application to Run

### **Deploy:**
```bash
pulumi up
```
### **Clean Up:**
```bash
pulumi destroy
```

This application will:
1. ✅ Ensure Secrets and Environment Variables are Properly Stored
2. ✅ Set up EFS mount targets 
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
- **Multi-AZ**: Deployed across 2 availability zones

### **4. Security Configuration**
- **ALB Security Group**: Allows HTTP/HTTPS from internet
- **Task Security Group**: Allows access only from ALB
- **EFS Security**: Transit encryption enabled

## 💰 **Cost Optimization**

- **EFS**: Provisioned throughput at 1 MiB/s (minimal cost)
- **Fargate**: 512 CPU units, 1024 MB memory
- **ALB**: Standard Application Load Balancer pricing
- **No NAT Gateway**: Uses public subnets with public IPs

## 🔧 **Environment Variables & Secrets for DB App**
1. **NEO4J_PASSWORD**: Generated Password for the Database, Sets the Database Password on Startup 

## 🔧 **Environment Variables & Secrets for Web App**
1. **DB_ADMIN**: Generated Password for the Database
2. **DB_URI**: Database Connection, should default to bolt://localhost:7687
3. **MODE**:"dev" for Development and "prod" for Production
4. **SESSION_KEY**: Authentication Key
5. **GMAIL_USER**: jrzesz@gmail.com should be default
6. **GMAIL_PASS**: Key for Email Access

## 🔧 **Backlog of Impovements**
1. : Generated Password for the Database