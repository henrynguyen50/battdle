#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "$REPO_ROOT"

echo -e "${BOLD}${BLUE}====================================================${NC}"
echo -e "${BOLD}${BLUE}   ⚾ PITCHLE PRE-COMMIT TEST SUITE RUNNER           ${NC}"
echo -e "${BOLD}${BLUE}====================================================${NC}\n"

FAILURES=0

# ---------------------------------------------------------
# Step 1: Frontend JavaScript & Kinematics Tests
# ---------------------------------------------------------
echo -e "${BOLD}▶ [1/3] Frontend JavaScript & Delivery Kinematics Tests...${NC}"
if node "${REPO_ROOT}/scripts/test_frontend.js"; then
    echo -e "${GREEN}✓ Frontend tests passed.${NC}\n"
else
    echo -e "${RED}✗ Frontend tests failed.${NC}\n"
    FAILURES=$((FAILURES + 1))
fi

# ---------------------------------------------------------
# Step 2: Go Backend Unit & Physics Tests
# ---------------------------------------------------------
echo -e "${BOLD}▶ [2/3] Go Services Unit & Kinematics Trajectory Tests...${NC}"
GO_CMD=""
if command -v go >/dev/null 2>&1; then
    GO_CMD="go test -v ./..."
else
    GO_CMD="docker run --rm -v ${REPO_ROOT}:/app -w /app golang:1.22-alpine go test -v ./..."
fi

if $GO_CMD; then
    echo -e "${GREEN}✓ Go backend tests passed.${NC}\n"
else
    echo -e "${RED}✗ Go backend tests failed.${NC}\n"
    FAILURES=$((FAILURES + 1))
fi

# ---------------------------------------------------------
# Step 3: PostgreSQL Schema & Database Integrity Tests
# ---------------------------------------------------------
echo -e "${BOLD}▶ [3/3] PostgreSQL Database Schema & Migration Tests...${NC}"
if bash "${REPO_ROOT}/scripts/test_postgres.sh"; then
    echo -e "${GREEN}✓ PostgreSQL database tests passed.${NC}\n"
else
    echo -e "${RED}✗ PostgreSQL database tests failed.${NC}\n"
    FAILURES=$((FAILURES + 1))
fi

# ---------------------------------------------------------
# Summary
# ---------------------------------------------------------
echo -e "${BOLD}${BLUE}====================================================${NC}"
if [ $FAILURES -eq 0 ]; then
    echo -e "${BOLD}${GREEN}  🎉 ALL PRE-COMMIT TESTS PASSED SUCCESSFULLY!       ${NC}"
    echo -e "${BOLD}${BLUE}====================================================${NC}"
    exit 0
else
    echo -e "${BOLD}${RED}  ❌ ${FAILURES} TEST SUITE(S) FAILED. COMMIT BLOCKED.     ${NC}"
    echo -e "${BOLD}${BLUE}====================================================${NC}"
    exit 1
fi
