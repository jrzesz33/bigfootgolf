#!/bin/bash

# BOAT Agent Local Development & Debug Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[BOAT]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Default values
MODE="test"
VERBOSE=false
CLAUDE_KEY=""
AWS_PROFILE="default"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--mode)
            MODE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -k|--claude-key)
            CLAUDE_KEY="$2"
            shift 2
            ;;
        -p|--aws-profile)
            AWS_PROFILE="$2"
            shift 2
            ;;
        -h|--help)
            echo "BOAT Agent Development & Debug Script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -m, --mode <mode>      Testing mode: test, interactive, sample, event"
            echo "  -v, --verbose          Enable verbose logging"
            echo "  -k, --claude-key <key> Claude API key"
            echo "  -p, --aws-profile <profile> AWS profile to use"
            echo "  -h, --help            Show this help message"
            echo ""
            echo "Modes:"
            echo "  test        Run predefined test scenarios"
            echo "  interactive Enter interactive testing mode"
            echo "  sample      Generate sample event files"
            echo "  event       Process specific event file (requires filename as next arg)"
            echo ""
            echo "Examples:"
            echo "  $0 --mode test --verbose"
            echo "  $0 --mode interactive --claude-key sk-xxx"
            echo "  $0 --mode event deploy-webapp.json"
            exit 0
            ;;
        *)
            EVENT_FILE="$1"
            shift
            ;;
    esac
done

print_status "Starting BOAT Agent local development environment"

# Check if we're in the right directory
if [ ! -f "go.mod" ] || ! grep -q "bigfoot/golf/opsagent" go.mod; then
    print_error "Please run this script from the opsagent directory"
    exit 1
fi

# Set up environment variables
export AWS_PROFILE="$AWS_PROFILE"

if [ -n "$CLAUDE_KEY" ]; then
    export CLAUDE_API_KEY="$CLAUDE_KEY"
    print_status "Using provided Claude API key"
elif [ -n "$CLAUDE_API_KEY" ]; then
    print_status "Using Claude API key from environment"
else
    print_warning "No Claude API key provided - using test mode"
    export CLAUDE_API_KEY="test-key-for-local-development"
fi

if [ "$VERBOSE" = true ]; then
    export LOG_LEVEL="debug"
    print_status "Verbose logging enabled"
fi

# Build the local development binary
print_status "Building BOAT agent for local development..."
if ! go build -o boat-local cmd/local/main.go; then
    print_error "Failed to build local development binary"
    exit 1
fi
print_success "Build complete"

# Check AWS credentials (optional for local testing)
if command -v aws >/dev/null 2>&1; then
    if aws sts get-caller-identity --profile "$AWS_PROFILE" >/dev/null 2>&1; then
        CALLER_ID=$(aws sts get-caller-identity --profile "$AWS_PROFILE" --query 'Account' --output text)
        print_status "AWS credentials valid for account: $CALLER_ID"
    else
        print_warning "AWS credentials not configured or invalid - some features may be limited"
    fi
else
    print_warning "AWS CLI not installed - some features may be limited"
fi

# Run the appropriate mode
print_status "Starting BOAT agent in '$MODE' mode"
echo ""

case $MODE in
    test)
        ./boat-local test
        ;;
    interactive)
        ./boat-local interactive
        ;;
    sample)
        ./boat-local sample
        ;;
    event)
        if [ -z "$EVENT_FILE" ]; then
            print_error "Event mode requires an event file"
            echo "Usage: $0 --mode event <event-file.json>"
            exit 1
        fi
        if [ ! -f "$EVENT_FILE" ]; then
            print_error "Event file not found: $EVENT_FILE"
            exit 1
        fi
        ./boat-local event "$EVENT_FILE"
        ;;
    *)
        print_error "Unknown mode: $MODE"
        echo "Valid modes: test, interactive, sample, event"
        exit 1
        ;;
esac

print_success "BOAT agent session completed"