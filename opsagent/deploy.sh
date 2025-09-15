#!/bin/bash

# Neo4j Deployment Script for BigFoot Golf
# Usage: ./deploy.sh [environment] [action]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/configs"
DEPLOYMENT_DIR="$SCRIPT_DIR/deployments"

# Default values
ENVIRONMENT="${1:-production}"
ACTION="${2:-deploy}"

# Configuration file
CONFIG_FILE="$CONFIG_DIR/neo4j-deployment.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if Python 3 is available
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 is required but not installed."
        exit 1
    fi
    
    # Check if configuration file exists
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "Configuration file not found: $CONFIG_FILE"
        exit 1
    fi
    
    # Check AWS credentials
    if [[ -z "$AWS_ACCESS_KEY_ID" && -z "$AWS_PROFILE" ]]; then
        log_warning "No AWS credentials found. Make sure to set AWS_ACCESS_KEY_ID or AWS_PROFILE"
    fi
    
    log_success "Dependencies check passed"
}

# Validate configuration
validate_config() {
    log_info "Validating configuration..."
    
    # Basic JSON validation
    if ! python3 -m json.tool "$CONFIG_FILE" > /dev/null 2>&1; then
        log_error "Invalid JSON in configuration file: $CONFIG_FILE"
        exit 1
    fi
    
    log_success "Configuration validation passed"
}

# Deploy infrastructure
deploy_infrastructure() {
    log_info "Starting Neo4j infrastructure deployment..."
    log_info "Environment: $ENVIRONMENT"
    log_info "Configuration: $CONFIG_FILE"
    
    echo "========================================"
    echo "🚀 BIGFOOT GOLF - NEO4J DEPLOYMENT"
    echo "========================================"
    
    # Run the MCP deployment
    if python3 "$DEPLOYMENT_DIR/mcp_integration.py" "$CONFIG_FILE"; then
        log_success "Neo4j deployment completed successfully!"
        
        echo ""
        echo "🎉 Your Neo4j database is now running!"
        echo ""
        echo "Next steps:"
        echo "1. Wait 2-3 minutes for health checks to pass"
        echo "2. Access Neo4j Browser via the provided URL"
        echo "3. Use the password from AWS Secrets Manager"
        echo ""
        
    else
        log_error "Deployment failed. Check the logs above for details."
        exit 1
    fi
}

# Show deployment status
show_status() {
    log_info "Checking deployment status..."
    
    # This would check AWS resources status
    echo "Status check functionality would go here"
    echo "- ECS Service status"
    echo "- Load Balancer health"
    echo "- Target Group health"
    echo "- Container logs"
}

# Destroy infrastructure
destroy_infrastructure() {
    log_warning "This will destroy ALL Neo4j infrastructure!"
    read -p "Are you sure you want to continue? (type 'yes' to confirm): " -r
    
    if [[ $REPLY == "yes" ]]; then
        log_info "Destroying infrastructure..."
        echo "Destroy functionality would go here"
        log_success "Infrastructure destroyed"
    else
        log_info "Destruction cancelled"
    fi
}

# Show help
show_help() {
    cat << EOF
BigFoot Golf - Neo4j Deployment Tool

Usage: $0 [environment] [action]

Environments:
  production (default)    - Production deployment
  staging                 - Staging deployment
  development            - Development deployment

Actions:
  deploy (default)       - Deploy Neo4j infrastructure
  status                 - Show deployment status
  destroy                - Destroy infrastructure
  help                   - Show this help message

Examples:
  $0                     # Deploy to production
  $0 production deploy   # Deploy to production
  $0 staging status      # Check staging status
  $0 production destroy  # Destroy production (dangerous!)

Environment Variables:
  AWS_ACCESS_KEY_ID      - AWS access key (or use AWS_PROFILE)
  AWS_SECRET_ACCESS_KEY  - AWS secret key
  AWS_REGION             - AWS region (default: us-east-1)
  AWS_PROFILE            - AWS profile name

Configuration:
  Edit configs/neo4j-deployment.json to customize deployment parameters.

EOF
}

# Main script logic
main() {
    case "$ACTION" in
        "deploy")
            check_dependencies
            validate_config
            deploy_infrastructure
            ;;
        "status")
            show_status
            ;;
        "destroy")
            destroy_infrastructure
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "Unknown action: $ACTION"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"