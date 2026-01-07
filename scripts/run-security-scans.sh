#!/bin/bash

# Security Scan Execution Script
# This script runs all local security scans for the Thums Up Backend

set -e

echo "🔒 Starting Security Scans for Thums Up Backend"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create reports directory
mkdir -p security-reports

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 1. gosec Scan
echo ""
echo -e "${YELLOW}[1/4] Running gosec security scan...${NC}"
if command_exists gosec; then
    gosec -fmt json -out security-reports/gosec-report.json ./... || true
    gosec -fmt sarif -out security-reports/gosec-results.sarif ./... || true
    echo -e "${GREEN}✓ gosec scan completed${NC}"
    echo "  Reports: security-reports/gosec-report.json"
    echo "           security-reports/gosec-results.sarif"
else
    echo -e "${RED}✗ gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest${NC}"
fi

# 2. govulncheck Scan
echo ""
echo -e "${YELLOW}[2/4] Running govulncheck vulnerability scan...${NC}"
if command_exists govulncheck; then
    govulncheck -json ./... > security-reports/govulncheck-report.json 2>&1 || true
    echo -e "${GREEN}✓ govulncheck scan completed${NC}"
    echo "  Report: security-reports/govulncheck-report.json"
    
    # Also show summary
    echo ""
    echo "Vulnerability Summary:"
    govulncheck ./... 2>&1 || true
else
    echo -e "${RED}✗ govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest${NC}"
fi

# 3. Dependency Check
echo ""
echo -e "${YELLOW}[3/4] Checking for outdated dependencies...${NC}"
if command_exists go; then
    echo "Checking for available updates..."
    go list -m -u all > security-reports/dependency-updates.txt 2>&1 || true
    echo -e "${GREEN}✓ Dependency check completed${NC}"
    echo "  Report: security-reports/dependency-updates.txt"
else
    echo -e "${RED}✗ Go not found${NC}"
fi

# 4. Secret Scanning (using git-secrets or similar)
echo ""
echo -e "${YELLOW}[4/4] Checking for potential secrets...${NC}"
echo "⚠️  Note: This is a basic check. Use GitHub Secret Scanning or GitGuardian for comprehensive scanning."

# Check for common secret patterns
SECRET_PATTERNS=(
    "password.*=.*['\"][^'\"]+['\"]"
    "api[_-]?key.*=.*['\"][^'\"]+['\"]"
    "secret.*=.*['\"][^'\"]+['\"]"
    "token.*=.*['\"][^'\"]+['\"]"
    "aws[_-]?access[_-]?key"
    "private[_-]?key"
)

FOUND_SECRETS=false
for pattern in "${SECRET_PATTERNS[@]}"; do
    if grep -r -i -E "$pattern" --include="*.go" --include="*.yaml" --include="*.yml" --include="*.json" \
        --exclude-dir=vendor --exclude-dir=node_modules --exclude-dir=.git . 2>/dev/null | \
        grep -v ".env.example" | grep -v "SECURITY" | grep -v "test" > /dev/null; then
        echo -e "${RED}⚠️  Potential secret pattern found: $pattern${NC}"
        FOUND_SECRETS=true
    fi
done

if [ "$FOUND_SECRETS" = false ]; then
    echo -e "${GREEN}✓ No obvious secrets found in code${NC}"
else
    echo -e "${YELLOW}⚠️  Review the codebase for potential secrets${NC}"
fi

# Summary
echo ""
echo "================================================"
echo -e "${GREEN}Security Scans Completed!${NC}"
echo ""
echo "Reports generated in: security-reports/"
echo ""
echo "Next steps:"
echo "1. Review all reports in security-reports/"
echo "2. Fix critical and high severity issues"
echo "3. Update SECURITY_ASSESSMENT.md with results"
echo "4. Run GitHub-based scans (CodeQL, Dependabot, Secret Scanning)"
echo "5. Run StackHawk DAST scan"
echo ""

