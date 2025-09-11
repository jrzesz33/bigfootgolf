#!/bin/bash

# BOAT Agent Full Functionality Demo
# Shows what happens with both AWS credentials and Claude API key

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              BOAT Agent Full Functionality Demo              ║${NC}"
echo -e "${BLUE}║                                                              ║${NC}"
echo -e "${BLUE}║  This demo shows what happens when you have BOTH:           ║${NC}"
echo -e "${BLUE}║  • AWS credentials                                           ║${NC}"
echo -e "${BLUE}║  • Claude API key                                            ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if we have credentials
HAS_AWS="no"
HAS_CLAUDE="no"

if aws sts get-caller-identity >/dev/null 2>&1; then
    HAS_AWS="yes"
    ACCOUNT_ID=$(aws sts get-caller-identity --query 'Account' --output text)
    echo -e "${GREEN}✅ AWS credentials found - Account: $ACCOUNT_ID${NC}"
else
    echo -e "${YELLOW}⚠️ No AWS credentials - will use simulation mode${NC}"
fi

if [ -n "$CLAUDE_API_KEY" ] && [ "$CLAUDE_API_KEY" != "test-key-for-local-development" ]; then
    HAS_CLAUDE="yes"
    echo -e "${GREEN}✅ Claude API key found - AI features enabled${NC}"
else
    echo -e "${YELLOW}⚠️ No Claude API key - will use fallback planning${NC}"
fi

echo ""

# Build if needed
if [ ! -f "build/boat-local" ]; then
    echo "Building BOAT agent..."
    make build-local >/dev/null 2>&1
    echo -e "${GREEN}✅ Build complete${NC}"
fi

echo ""
echo -e "${CYAN}🎭 Demo Scenarios:${NC}"
echo ""

# Scenario 1: Simple deployment
echo -e "${YELLOW}1. Simple Web Application Deployment${NC}"
echo "   Task: Deploy a React application with load balancer"
echo ""

cat > demo-simple.json << 'EOF'
{
  "id": "demo-simple-001",
  "type": "deploy",
  "description": "Deploy a React web application with nginx and load balancer for a small startup",
  "parameters": {
    "image": "ghcr.io/jrzesz33/react-app:v1.2.0",
    "expected_users": "100-500",
    "budget": "minimal"
  },
  "priority": "medium",
  "requester": "startup-team"
}
EOF

echo "Running scenario..."
if ./build/boat-local event demo-simple.json; then
    echo -e "${GREEN}✅ Simple deployment scenario completed${NC}"
else
    echo -e "${YELLOW}⚠️ Scenario completed with warnings${NC}"
fi

echo ""
echo "───────────────────────────────────────────────────────────────"
echo ""

# Scenario 2: Complex deployment
echo -e "${YELLOW}2. Complex Microservices Deployment${NC}"
echo "   Task: Deploy a complex microservices architecture"
echo ""

cat > demo-complex.json << 'EOF'
{
  "id": "demo-complex-001", 
  "type": "deploy",
  "description": "Deploy a complex microservices architecture with API gateway, authentication service, user management, payment processing, Redis cache, and PostgreSQL database. Needs to handle 10,000+ concurrent users with high availability and auto-scaling.",
  "parameters": {
    "services": ["api-gateway", "auth-service", "user-service", "payment-service"],
    "images": [
      "ghcr.io/jrzesz33/api-gateway:v2.1.0",
      "ghcr.io/jrzesz33/auth-service:v1.5.0", 
      "ghcr.io/jrzesz33/user-service:v3.0.1",
      "ghcr.io/jrzesz33/payment-service:v1.8.2"
    ],
    "expected_users": "10000+",
    "high_availability": true,
    "auto_scaling": true,
    "monitoring": "comprehensive",
    "budget": "moderate"
  },
  "priority": "high",
  "requester": "enterprise-team"
}
EOF

echo "Running complex scenario..."
if ./build/boat-local event demo-complex.json; then
    echo -e "${GREEN}✅ Complex deployment scenario completed${NC}"
else
    echo -e "${YELLOW}⚠️ Scenario completed with warnings${NC}"
fi

echo ""
echo "───────────────────────────────────────────────────────────────"
echo ""

# Scenario 3: Troubleshooting
echo -e "${YELLOW}3. Troubleshooting Scenario${NC}"
echo "   Task: Diagnose and fix a production issue"
echo ""

cat > demo-troubleshoot.json << 'EOF'
{
  "id": "demo-troubleshoot-001",
  "type": "troubleshoot", 
  "description": "Production API is experiencing intermittent 500 errors, high response times (3-5 seconds), and memory usage is at 90%. Users are complaining about timeouts. Database queries seem slow and Redis cache hit rate has dropped to 60%.",
  "parameters": {
    "service": "user-api-service",
    "symptoms": [
      "intermittent 500 errors (15% error rate)",
      "high response times (3-5 seconds avg)",
      "memory usage at 90%", 
      "slow database queries",
      "low cache hit rate (60%)"
    ],
    "impact": "high",
    "affected_users": "all users",
    "priority": "urgent"
  },
  "priority": "urgent",
  "requester": "on-call-engineer"
}
EOF

echo "Running troubleshooting scenario..."
if ./build/boat-local event demo-troubleshoot.json; then
    echo -e "${GREEN}✅ Troubleshooting scenario completed${NC}"
else
    echo -e "${YELLOW}⚠️ Scenario completed with warnings${NC}"
fi

# Cleanup
rm -f demo-*.json

echo ""
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                        Demo Summary                          ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${CYAN}What You Just Saw:${NC}"
echo ""

if [ "$HAS_CLAUDE" = "yes" ]; then
    echo -e "${GREEN}🧠 Claude AI Analysis:${NC}"
    echo "  • Natural language task processing"
    echo "  • Intelligent deployment planning" 
    echo "  • Cost optimization suggestions"
    echo "  • Resource dependency analysis"
    echo "  • Troubleshooting recommendations"
else
    echo -e "${YELLOW}🤖 Fallback Planning:${NC}"
    echo "  • Basic deployment templates"
    echo "  • Standard cost estimates"
    echo "  • Rule-based resource planning"
fi

echo ""

if [ "$HAS_AWS" = "yes" ]; then
    echo -e "${GREEN}☁️ AWS Integration:${NC}"
    echo "  • Real VPC and subnet discovery"
    echo "  • Actual ECS cluster information"
    echo "  • Live cost calculations"
    echo "  • Resource creation capabilities"
else
    echo -e "${YELLOW}☁️ AWS Simulation:${NC}"
    echo "  • Simulated resource discovery"
    echo "  • Estimated cost calculations"
    echo "  • Validation without actual deployment"
fi

echo ""
echo -e "${GREEN}🎯 Always Working Features:${NC}"
echo "  • Container requirement validation"
echo "  • Cost threshold monitoring"
echo "  • Registry compliance checking (ghcr.io/jrzesz33/)"
echo "  • Free tier compliance validation"
echo "  • Structured logging and monitoring"
echo "  • Error handling and recovery"

echo ""
echo -e "${CYAN}🚀 Next Steps:${NC}"
echo ""

if [ "$HAS_AWS" = "yes" ] && [ "$HAS_CLAUDE" = "yes" ]; then
    echo -e "${GREEN}You have FULL functionality! Ready for:${NC}"
    echo "• Production deployments"
    echo "• Complex multi-service architectures"
    echo "• AI-powered troubleshooting"
    echo "• Cost-optimized infrastructure planning"
    echo ""
    echo "Try interactive mode: make dev-interactive"
elif [ "$HAS_AWS" = "yes" ] || [ "$HAS_CLAUDE" = "yes" ]; then
    echo -e "${YELLOW}You have PARTIAL functionality:${NC}"
    if [ "$HAS_AWS" = "no" ]; then
        echo "• Add AWS credentials for real infrastructure deployment"
    fi
    if [ "$HAS_CLAUDE" = "no" ]; then
        echo "• Add Claude API key for AI-powered planning"
    fi
    echo ""
    echo "Run: ./setup-full-integration.sh"
else
    echo -e "${YELLOW}Ready for development and testing:${NC}"
    echo "• All validation logic works"
    echo "• Cost estimation works"
    echo "• Container requirement checking works"
    echo ""
    echo "For full functionality, run: ./setup-full-integration.sh"
fi

echo ""
echo -e "${BLUE}Demo completed! Check the logs above to see how BOAT analyzed each scenario.${NC}"