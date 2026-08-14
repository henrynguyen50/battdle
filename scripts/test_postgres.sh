#!/bin/bash
set -e

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🐘 Running PostgreSQL Database Precommit Tests...${NC}\n"

# Determine how to execute psql (local psql or docker exec)
DB_CONTAINER="pitchle_db"
DB_USER="postgres"
DB_NAME="pitchle"
TEST_DB="pitchle_precommit_test"

run_sql() {
    local database="$1"
    local sql_query="$2"

    if command -v psql >/dev/null 2>&1 && pg_isready -h localhost -p 5432 -U "$DB_USER" >/dev/null 2>&1; then
        PGPASSWORD=postgres psql -h localhost -p 5432 -U "$DB_USER" -d "$database" -t -A -c "$sql_query" 2>&1
    elif docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER}$"; then
        docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$database" -t -A -c "$sql_query" 2>&1
    else
        echo "NO_DB"
    fi
}

run_sql_file() {
    local database="$1"
    local sql_file="$2"

    if command -v psql >/dev/null 2>&1 && pg_isready -h localhost -p 5432 -U "$DB_USER" >/dev/null 2>&1; then
        PGPASSWORD=postgres psql -h localhost -p 5432 -U "$DB_USER" -d "$database" -f "$sql_file" >/dev/null 2>&1
    elif docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER}$"; then
        docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$database" < "$sql_file" >/dev/null 2>&1
    else
        echo "NO_DB"
    fi
}

# 1. Check DB Connection
echo "1. Checking Postgres Database Availability:"
CHECK=$(run_sql "postgres" "SELECT 1;")
if [ "$CHECK" = "NO_DB" ]; then
    echo -e "${YELLOW}  ⚠ Warning: PostgreSQL is not running locally or in Docker container '${DB_CONTAINER}'. Starting via docker compose...${NC}"
    docker compose up -d db
    sleep 3
    CHECK=$(run_sql "postgres" "SELECT 1;")
fi

if [ "$CHECK" != "1" ]; then
    echo -e "${RED}  ❌ FAIL: Unable to connect to PostgreSQL.${NC}"
    exit 1
fi
echo -e "${GREEN}  ✓ PASS: PostgreSQL is connected and healthy.${NC}"

# 2. Test Ephemeral Schema Migration (Up & Down)
echo -e "\n2. Testing Schema Migrations on Ephemeral Test Database ('${TEST_DB}'):"

# Clean up old test db if exists
run_sql "postgres" "DROP DATABASE IF EXISTS ${TEST_DB};" >/dev/null 2>&1
run_sql "postgres" "CREATE DATABASE ${TEST_DB};" >/dev/null 2>&1

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

UP_MIGRATION="${REPO_ROOT}/database/migrations/000001_init_schema.up.sql"
DOWN_MIGRATION="${REPO_ROOT}/database/migrations/000001_init_schema.down.sql"

if [ ! -f "$UP_MIGRATION" ] || [ ! -f "$DOWN_MIGRATION" ]; then
    echo -e "${RED}  ❌ FAIL: Migration files not found in database/migrations/${NC}"
    run_sql "postgres" "DROP DATABASE IF EXISTS ${TEST_DB};" >/dev/null 2>&1
    exit 1
fi

# Run Up Migration
run_sql_file "${TEST_DB}" "$UP_MIGRATION"
echo -e "${GREEN}  ✓ PASS: 000001_init_schema.up.sql executed successfully.${NC}"

# Verify created tables count
TABLE_COUNT=$(run_sql "${TEST_DB}" "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';")
if [ "$TABLE_COUNT" -ge 8 ]; then
    echo -e "${GREEN}  ✓ PASS: All 8 core tables created in test database (count=${TABLE_COUNT}).${NC}"
else
    echo -e "${RED}  ❌ FAIL: Expected at least 8 tables, found ${TABLE_COUNT}.${NC}"
    run_sql "postgres" "DROP DATABASE IF EXISTS ${TEST_DB};" >/dev/null 2>&1
    exit 1
fi

# Verify specific table columns and constraints
DIV_COL_CHECK=$(run_sql "${TEST_DB}" "SELECT count(*) FROM information_schema.columns WHERE table_name = 'pitch_profiles' AND column_name IN ('arm_angle', 'release_extension', 'break_x', 'break_z');")
if [ "$DIV_COL_CHECK" -eq 4 ]; then
    echo -e "${GREEN}  ✓ PASS: pitch_profiles schema contains required Statcast kinematics columns.${NC}"
else
    echo -e "${RED}  ❌ FAIL: pitch_profiles missing required Statcast kinematics columns.${NC}"
    run_sql "postgres" "DROP DATABASE IF EXISTS ${TEST_DB};" >/dev/null 2>&1
    exit 1
fi

# Run Down Migration
run_sql_file "${TEST_DB}" "$DOWN_MIGRATION"
echo -e "${GREEN}  ✓ PASS: 000001_init_schema.down.sql executed successfully.${NC}"

DOWN_TABLE_COUNT=$(run_sql "${TEST_DB}" "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';")
if [ "$DOWN_TABLE_COUNT" -eq 0 ]; then
    echo -e "${GREEN}  ✓ PASS: All tables dropped cleanly during rollback.${NC}"
else
    echo -e "${RED}  ❌ FAIL: Expected 0 tables after down migration, found ${DOWN_TABLE_COUNT}.${NC}"
    run_sql "postgres" "DROP DATABASE IF EXISTS ${TEST_DB};" >/dev/null 2>&1
    exit 1
fi

# Teardown test database
run_sql "postgres" "DROP DATABASE IF EXISTS ${TEST_DB};" >/dev/null 2>&1
echo -e "${GREEN}  ✓ PASS: Ephemeral test database cleaned up.${NC}"

# 3. Validate Live Application Database Integrity (pitchle)
echo -e "\n3. Validating Live Database Integrity ('${DB_NAME}'):"

ORPHAN_PLAYERS=$(run_sql "${DB_NAME}" "SELECT count(*) FROM players WHERE team_id IS NOT NULL AND team_id NOT IN (SELECT id FROM teams);")
if [ "$ORPHAN_PLAYERS" = "0" ]; then
    echo -e "${GREEN}  ✓ PASS: Zero orphaned player-team foreign keys.${NC}"
else
    echo -e "${RED}  ❌ FAIL: Found ${ORPHAN_PLAYERS} orphaned player foreign keys.${NC}"
    exit 1
fi

ORPHAN_PROFILES=$(run_sql "${DB_NAME}" "SELECT count(*) FROM pitch_profiles WHERE player_id NOT IN (SELECT id FROM players);")
if [ "$ORPHAN_PROFILES" = "0" ]; then
    echo -e "${GREEN}  ✓ PASS: Zero orphaned pitch_profiles foreign keys.${NC}"
else
    echo -e "${RED}  ❌ FAIL: Found ${ORPHAN_PROFILES} orphaned pitch profiles.${NC}"
    exit 1
fi

TOTAL_PITCHERS=$(run_sql "${DB_NAME}" "SELECT count(*) FROM players;")
TOTAL_PROFILES=$(run_sql "${DB_NAME}" "SELECT count(*) FROM pitch_profiles;")
echo -e "${GREEN}  ✓ PASS: Verified database data (${TOTAL_PITCHERS} pitchers, ${TOTAL_PROFILES} pitch profiles).${NC}"

echo -e "\n========================================"
echo -e "${GREEN}Postgres Database Tests: ALL PASSED${NC}"
echo -e "========================================"
exit 0
