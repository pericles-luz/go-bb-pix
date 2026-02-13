#!/bin/bash

set -e

echo "================================================"
echo "  go-bb-pix Test Suite"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
        exit 1
    fi
}

print_section() {
    echo ""
    echo -e "${YELLOW}▶ $1${NC}"
    echo "------------------------------------------------"
}

# 1. Run unit tests
print_section "Running Unit Tests"
go test ./... -short
print_status $? "Unit tests passed"

# 2. Run validation tests
print_section "Running Validation Tests"
go test ./pix -v -run TestValidation -short > /dev/null 2>&1
print_status $? "Validation tests passed"

# 3. Run integration tests
print_section "Running Integration Tests"
go test ./pix -v -run TestIntegration -short > /dev/null 2>&1
print_status $? "Integration tests passed"

# 4. Run error handling tests
print_section "Running Error Handling Tests"
go test ./pix -v -run TestAPIErrorResponses -short > /dev/null 2>&1
go test ./pix -v -run TestHTTPStatusCodes -short > /dev/null 2>&1
print_status $? "Error handling tests passed"

# 5. Run CobV tests
print_section "Running CobV Tests"
go test ./pix -v -run TestCobV -short > /dev/null 2>&1
print_status $? "CobV tests passed"

# 6. Run webhook tests
print_section "Running Webhook Tests"
go test ./pix -v -run TestWebhook -short > /dev/null 2>&1
print_status $? "Webhook tests passed"

# 7. Generate coverage report
print_section "Generating Coverage Report"
go test ./... -short -coverprofile=coverage.out > /dev/null 2>&1
go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: "$3}'
print_status $? "Coverage report generated"

# 8. Show package coverage
echo ""
echo "Coverage by Package:"
go test ./... -short -cover 2>&1 | grep -E "^ok" | awk '{print "  " $2 ": " $5}'

# 9. Count tests
echo ""
echo "Test Statistics:"
TEST_FILES=$(find . -name "*_test.go" -type f | wc -l)
echo "  Test files: $TEST_FILES"

TEST_CASES=$(go test ./pix -short -v 2>&1 | grep -E "^=== RUN" | wc -l)
echo "  Test cases in pix package: $TEST_CASES"

# 10. Check for E2E tests availability
echo ""
if [ -n "$BB_CLIENT_ID" ] && [ -n "$BB_CLIENT_SECRET" ] && [ -n "$BB_DEV_APP_KEY" ]; then
    echo -e "${GREEN}E2E tests available${NC} (credentials configured)"
    echo "  Run with: go test -v -tags=integration ./pix"
else
    echo -e "${YELLOW}E2E tests not available${NC} (credentials not configured)"
    echo "  Configure: BB_CLIENT_ID, BB_CLIENT_SECRET, BB_DEV_APP_KEY"
fi

echo ""
echo "================================================"
echo -e "${GREEN}✓ All tests passed!${NC}"
echo "================================================"
echo ""
echo "For more details, see:"
echo "  - TESTING.md for testing guide"
echo "  - TEST_SUMMARY.md for test summary"
echo "  - testdata/README.md for fixture documentation"
echo ""
