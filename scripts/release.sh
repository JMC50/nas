#!/usr/bin/env bash
# release.sh — Bump version, commit, tag, and push.
#
# Usage:
#   scripts/release.sh <new_version> [--dry-run]
#
# Arguments:
#   new_version   target version, semver form (e.g. 0.0.2-beta).
#                 The leading "v" is added automatically for the git tag.
#   --dry-run     print actions without modifying files or running git.
#
# Required state:
#   - Clean working tree (no uncommitted changes).
#   - On `main` branch.
#   - Docs/release-notes/v<new_version>.md exists and is polished.
#     Generate the initial draft via scripts/release-draft.sh first.
#
# Actions:
#   1. Validate semver, branch, working tree, tag uniqueness, note file presence.
#   2. Write <new_version> to VERSION (single source of truth).
#   3. Sync package.json (root) + frontend/package.json `version` fields so all
#      three sources stay locked together. The root package.json is a Node-era
#      artifact that still pins frontend devDependencies (vite, svelte-kit,
#      tailwindcss) used by the Docker build; until those move into
#      frontend/package.json its `version` field stays in lockstep here.
#   4. Stage VERSION, both package.json files, the polished notes, commit
#      `[chore] release v<new_version>` (skip if dry-run).
#   5. Create annotated tag `v<new_version>`.
#   6. Push commit + tag to origin/main (skip if dry-run).

set -euo pipefail

NEW_VERSION="${1:-}"
DRY_RUN=false
for arg in "$@"; do
  [[ "$arg" == "--dry-run" ]] && DRY_RUN=true
done

if [[ -z "$NEW_VERSION" || "$NEW_VERSION" == "--dry-run" ]]; then
  echo "usage: $0 <new_version> [--dry-run]" >&2
  echo "       version must be semver (e.g. 0.0.2-beta, 1.2.0)" >&2
  exit 2
fi

# Reject the leading "v" — script appends it for the tag.
if [[ "$NEW_VERSION" =~ ^v ]]; then
  echo "error: pass version WITHOUT leading 'v' (got: $NEW_VERSION)" >&2
  exit 2
fi

# Loose semver check: M.m.p with optional -prerelease[.N] / +build.
if ! [[ "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
  echo "error: '$NEW_VERSION' is not a valid semver string" >&2
  exit 2
fi

TAG="v${NEW_VERSION}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Pre-flight checks.
branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$branch" != "main" ]]; then
  echo "error: must be on 'main' branch (currently: $branch)" >&2
  exit 2
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "error: working tree has uncommitted changes:" >&2
  git status --short >&2
  exit 2
fi

if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "error: tag '$TAG' already exists" >&2
  exit 2
fi

NOTE_FILE="Docs/release-notes/${TAG}.md"
NOTE_FILE_KO="Docs/release-notes/${TAG}.ko.md"
if [[ ! -f "$NOTE_FILE" ]]; then
  echo "error: English release note not found at $NOTE_FILE" >&2
  echo "       run: scripts/release-draft.sh $NEW_VERSION" >&2
  echo "       then polish the generated draft (English main + Korean translation) before re-running this script." >&2
  exit 2
fi
if [[ ! -f "$NOTE_FILE_KO" ]]; then
  echo "error: Korean release note not found at $NOTE_FILE_KO" >&2
  echo "       this project ships bilingual notes — both must exist before tagging." >&2
  exit 2
fi

run() {
  if $DRY_RUN; then
    echo "[dry-run] $*"
  else
    "$@"
  fi
}

write_file() {
  local path="$1"
  local content="$2"
  if $DRY_RUN; then
    echo "[dry-run] write '$path' <- '$content'"
  else
    printf '%s\n' "$content" > "$path"
  fi
}

echo "=== Release pipeline: $TAG ==="
echo "  Repo:    $REPO_ROOT"
echo "  Branch:  $branch"
echo "  Note:    $NOTE_FILE"
echo "  Dry-run: $DRY_RUN"
echo

# 1. VERSION file.
echo "[1/5] write VERSION"
write_file "VERSION" "$NEW_VERSION"

# 2. package.json (root) + frontend/package.json — match first `"version"` line
#    in each, preserve formatting. Cross-platform sed.
sync_pkg_version() {
  local path="$1"
  if $DRY_RUN; then
    echo "[dry-run] sed -i 's/\"version\": .*/\"version\": \"$NEW_VERSION\",/' $path"
  else
    if sed --version >/dev/null 2>&1; then
      sed -i "0,/\"version\": .*/{s//\"version\": \"$NEW_VERSION\",/}" "$path"
    else
      sed -i '' "0,/\"version\": .*/{s//\"version\": \"$NEW_VERSION\",/;}" "$path"
    fi
  fi
}
echo "[2/5] sync package.json (root + frontend) version"
sync_pkg_version package.json
sync_pkg_version frontend/package.json

# 3. Stage + commit.
echo "[3/5] commit"
run git add VERSION package.json frontend/package.json "$NOTE_FILE" "$NOTE_FILE_KO"
run git commit -m "[chore] release ${TAG}"

# 4. Annotated tag pointing at the release commit. Body of the tag annotation is the note file.
echo "[4/5] tag ${TAG}"
if $DRY_RUN; then
  echo "[dry-run] git tag -a ${TAG} -F ${NOTE_FILE}"
else
  git tag -a "$TAG" -F "$NOTE_FILE"
fi

# 5. Push commit + tag.
echo "[5/5] push origin main --follow-tags"
run git push origin main --follow-tags

echo
echo "Done. Tag pushed: $TAG"
echo "  GitHub Release: if .github/workflows/release.yml is enabled, it will create the release automatically."
echo "  Manual fallback: gh release create $TAG --notes-file $NOTE_FILE"
