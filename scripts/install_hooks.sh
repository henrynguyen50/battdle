#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "Setting up Git pre-commit hooks..."

# Option A: Configure Git core.hooksPath to .githooks
git -C "$REPO_ROOT" config core.hooksPath .githooks

# Option B: Also copy to .git/hooks/pre-commit as fallback
if [ -d "${REPO_ROOT}/.git/hooks" ]; then
    cp "${REPO_ROOT}/.githooks/pre-commit" "${REPO_ROOT}/.git/hooks/pre-commit"
    chmod +x "${REPO_ROOT}/.git/hooks/pre-commit"
fi

chmod +x "${REPO_ROOT}/.githooks/pre-commit"
chmod +x "${REPO_ROOT}/scripts/run_all_tests.sh"
chmod +x "${REPO_ROOT}/scripts/test_postgres.sh"

echo "✓ Pre-commit hooks successfully installed and active!"
