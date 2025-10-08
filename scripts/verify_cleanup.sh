#!/bin/bash
# Verification script for code cleanup changes

set -e

echo "=================================="
echo "Golf App Cleanup Verification"
echo "=================================="
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
PASSED=0
FAILED=0
WARNINGS=0

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED++))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED++))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS++))
}

echo "1. Checking for critical files..."
if [ -f ".env.example" ]; then
    check_pass ".env.example exists"
else
    check_fail ".env.example missing"
fi

if [ -f "pkg/helper/config.go" ]; then
    check_pass "pkg/helper/config.go exists"
else
    check_fail "pkg/helper/config.go missing"
fi

if [ -f "pkg/helper/constants.go" ]; then
    check_pass "pkg/helper/constants.go exists"
else
    check_fail "pkg/helper/constants.go missing"
fi

if [ -f ".golangci.yml" ]; then
    check_pass ".golangci.yml exists"
else
    check_fail ".golangci.yml missing"
fi

if [ -f "TODO.md" ]; then
    check_pass "TODO.md exists"
else
    check_fail "TODO.md missing"
fi

echo ""
echo "2. Checking for fixed bugs..."

# Check for birdsfoot bug
if grep -r "birdsfoot" pkg/ 2>/dev/null | grep -v "Binary"; then
    check_fail "Found 'birdsfoot' references (should be 'bigfoot')"
else
    check_pass "No 'birdsfoot' references found (bug fixed)"
fi

echo ""
echo "3. Checking .gitignore patterns..."

if grep -q "__debug_bin" .gitignore; then
    check_pass ".gitignore includes __debug_bin pattern"
else
    check_fail ".gitignore missing __debug_bin pattern"
fi

if grep -q "\.wasm" .gitignore; then
    check_pass ".gitignore includes .wasm pattern"
else
    check_fail ".gitignore missing .wasm pattern"
fi

echo ""
echo "4. Checking environment variable documentation..."

ENV_VARS=("DB_ADMIN" "JWT_SECRET" "LLM_GATEWAY_URL" "MCP_GATEWAY_URL")
for var in "${ENV_VARS[@]}"; do
    if grep -q "$var" .env.example; then
        check_pass "$var documented in .env.example"
    else
        check_fail "$var missing from .env.example"
    fi
done

echo ""
echo "5. Checking Go workspace configuration..."

if [ -f "go.work" ]; then
    check_pass "go.work exists"
else
    check_warn "go.work missing (should exist for workspace)"
fi

echo ""
echo "6. Checking package documentation..."

PACKAGES=("pkg/models/db/db.go" "pkg/models/auth/init.go" "pkg/handlers/auth.go")
for pkg in "${PACKAGES[@]}"; do
    if [ -f "$pkg" ]; then
        if head -20 "$pkg" | grep -q "^// Package"; then
            check_pass "$pkg has package documentation"
        else
            check_warn "$pkg missing package documentation"
        fi
    fi
done

echo ""
echo "7. Testing Go workspace sync..."

if command -v go &> /dev/null; then
    if go work sync 2>&1 | grep -q "error"; then
        check_fail "go work sync failed"
    else
        check_pass "go work sync successful"
    fi
else
    check_warn "Go not installed, skipping workspace test"
fi

echo ""
echo "8. Checking for WASM build path in README..."

if grep -q "web/public/app.wasm" README.md; then
    check_pass "README uses correct WASM path"
else
    check_fail "README has incorrect WASM path"
fi

if grep -q "web/main.go" README.md; then
    check_pass "README uses correct entry point"
else
    check_fail "README references wrong entry point"
fi

echo ""
echo "9. Checking for commented code (should be removed)..."

COMMENTED_FILES=("pkg/models/db/helper.go" "web/main.go")
FOUND_COMMENTS=false
for file in "${COMMENTED_FILES[@]}"; do
    if [ -f "$file" ]; then
        # Check for block comments that look like commented code
        if grep -E "^/\*$" "$file" | head -5 | grep -q "func\|if\|for"; then
            check_fail "$file still has commented code blocks"
            FOUND_COMMENTS=true
        fi
    fi
done

if [ "$FOUND_COMMENTS" = false ]; then
    check_pass "No commented code blocks found"
fi

echo ""
echo "10. Verifying constants usage..."

if grep -q "DefaultDBPoolSize" pkg/models/db/db.go; then
    check_pass "Database uses constants from helper package"
else
    check_warn "Database may not be using constants"
fi

if grep -q "GoogleTokenURL" pkg/models/auth/init.go; then
    check_pass "Auth uses OAuth URL constants"
else
    check_warn "Auth may not be using OAuth constants"
fi

echo ""
echo "=================================="
echo "Verification Summary"
echo "=================================="
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${YELLOW}Warnings:${NC} $WARNINGS"
echo -e "${RED}Failed:${NC} $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All critical checks passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some checks failed. Please review the output above.${NC}"
    exit 1
fi
