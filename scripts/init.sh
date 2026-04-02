#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# init.sh — Initialize a new project from the go-ygg template
#
# Usage: ./scripts/init.sh <project-name>
#
# Example: ./scripts/init.sh my-app
# ---------------------------------------------------------------------------

NEW_KEBAB="${1:-}"

if [[ -z "$NEW_KEBAB" ]]; then
  echo "Usage: $0 <project-name>" >&2
  echo "Example: $0 my-app" >&2
  exit 1
fi

if [[ ! "$NEW_KEBAB" =~ ^[a-z][a-z0-9-]*$ ]]; then
  echo "Error: project name must be lowercase, start with a letter, and contain only letters, digits, and hyphens." >&2
  exit 1
fi

# Derive SCREAMING_SNAKE_CASE from kebab-case
NEW_SCREAMING="${NEW_KEBAB//-/_}"
NEW_SCREAMING="${NEW_SCREAMING^^}"

# ---------------------------------------------------------------------------
# Cross-platform sed -i (Linux vs macOS)
# ---------------------------------------------------------------------------
sed_inplace() {
  if [[ "$(uname)" == "Darwin" ]]; then
    sed -i '' "$@"
  else
    sed -i "$@"
  fi
}

# ---------------------------------------------------------------------------
# Replace in a file: GO_YGG_ → NEW_SCREAMING_, GO_YGG → NEW_SCREAMING, go-ygg → NEW_KEBAB
# Order matters: most specific pattern first.
# ---------------------------------------------------------------------------
rename_in_file() {
  local file="$1"
  [[ -f "$file" ]] || return 0
  sed_inplace \
    -e "s|GO_YGG_|${NEW_SCREAMING}_|g" \
    -e "s|GO_YGG|${NEW_SCREAMING}|g" \
    -e "s|go-ygg|${NEW_KEBAB}|g" \
    "$file"
}

# ---------------------------------------------------------------------------
# Files to update
# ---------------------------------------------------------------------------
FILES=(
  go.mod
  Makefile
  Dockerfile
  docker-compose.yml
  .golangci.yml
  config.example.yaml
  README.md
  docs/technical/database.md
  internal/infrastructure/config/config.go
  internal/interfaces/cli/root.go
  internal/interfaces/cli/serve.go
  internal/interfaces/cli/migrate.go
  internal/interfaces/cli/version.go
  internal/interfaces/http/router.go
  internal/interfaces/http/handler/home.go
  internal/interfaces/http/handler/version.go
  internal/application/example/usecase.go
  internal/infrastructure/persistence/example_repository.go
  cmd/main.go
)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$SCRIPT_DIR")"

for file in "${FILES[@]}"; do
  rename_in_file "$ROOT/$file"
done

echo "✓ Renamed go-ygg → $NEW_KEBAB (GO_YGG → $NEW_SCREAMING)"
echo ""
echo "Next steps:"
echo "  go mod tidy"
echo "  git add -A && git commit -m \"chore: initialize project as $NEW_KEBAB\""
