#!/usr/bin/env bash
# release-draft.sh — Generate a mechanical release-notes draft from git log.
#
# Usage:
#   scripts/release-draft.sh <new_version> [previous_ref]
#
# Arguments:
#   new_version    target version (semver, e.g. 0.0.2-beta). Used in output filename + header.
#   previous_ref   optional. Range start. Defaults to the most recent annotated tag
#                  matching v*, or the repo's root commit if no tag exists.
#
# Output:
#   Docs/release-notes/v<new_version>.md
#
# Behaviour:
#   - Reads git log <previous_ref>..HEAD with --no-merges.
#   - Parses `[bracket] subject` prefix and groups commits by category:
#       Features      ← [feat]
#       Fixes         ← [fix], [!HOTFIX]
#       UI Refinements← [design], [style]
#       Infrastructure← [refactor], [chore], [rename], [remove], [comment]
#       Documentation ← [docs]
#       Tests         ← [test]
#       Other         ← anything else (incl. messages without bracket prefix)
#   - Breaking commits ([!BREAKING CHANGE]) bubble to a separate top section.
#   - Writes a markdown draft with section scaffolding the author still has to polish.
#
# This script does NOT bump versions, tag, or push. Use scripts/release.sh for that.

set -euo pipefail

NEW_VERSION="${1:-}"
PREV_REF="${2:-}"

if [[ -z "$NEW_VERSION" ]]; then
  echo "usage: $0 <new_version> [previous_ref]" >&2
  exit 2
fi

# Resolve repo root regardless of where this script was invoked from.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Pick previous ref: explicit arg > most recent v* tag > first commit.
if [[ -z "$PREV_REF" ]]; then
  PREV_REF="$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || true)"
  if [[ -z "$PREV_REF" ]]; then
    PREV_REF="$(git rev-list --max-parents=0 HEAD | tail -n1)"
  fi
fi

RANGE="${PREV_REF}..HEAD"
OUTPUT_DIR="Docs/release-notes"
OUTPUT_FILE="${OUTPUT_DIR}/v${NEW_VERSION}.md"
mkdir -p "$OUTPUT_DIR"

# Pull commit pairs. Format: <short_hash>|<subject>
# --no-merges drops merge commits (individual commits hold the real change content).
mapfile -t LINES < <(git log "$RANGE" --no-merges --pretty=format:'%h|%s')

# Bucket commits by category. Bash 4+ associative arrays.
declare -A BUCKETS
BUCKETS[breaking]=""
BUCKETS[features]=""
BUCKETS[fixes]=""
BUCKETS[ui]=""
BUCKETS[infra]=""
BUCKETS[docs]=""
BUCKETS[tests]=""
BUCKETS[other]=""

for line in "${LINES[@]}"; do
  [[ -z "$line" ]] && continue
  hash="${line%%|*}"
  subject="${line#*|}"

  # Match bracket prefix: [type] or [!TYPE]
  if [[ "$subject" =~ ^\[!?([A-Za-z!\ ]+)\][[:space:]]+(.+)$ ]]; then
    raw_type="${BASH_REMATCH[1]}"
    body="${BASH_REMATCH[2]}"
    type="$(echo "$raw_type" | tr '[:upper:]' '[:lower:]' | tr -d '! ')"
  else
    type=""
    body="$subject"
  fi

  bullet="- \`${hash}\` ${body}"$'\n'

  case "$type" in
    breakingchange|hotfix) BUCKETS[breaking]+="$bullet" ;;
    feat)                  BUCKETS[features]+="$bullet" ;;
    fix)                   BUCKETS[fixes]+="$bullet" ;;
    design|style)          BUCKETS[ui]+="$bullet" ;;
    refactor|chore|rename|remove|comment)
                           BUCKETS[infra]+="$bullet" ;;
    docs)                  BUCKETS[docs]+="$bullet" ;;
    test)                  BUCKETS[tests]+="$bullet" ;;
    *)                     BUCKETS[other]+="$bullet" ;;
  esac
done

# Count for header table.
total=$(( ${#LINES[@]} ))
merges=$(git log "$RANGE" --merges --oneline | wc -l | tr -d '[:space:]')

# Emit draft. The author replaces bullet lists with polished prose per the
# release-notes skill rules; this file is just the raw material.
{
  echo "# Release Notes — v${NEW_VERSION}"
  echo
  echo "| | |"
  echo "|---|---|"
  echo "| Range | \`${PREV_REF}..HEAD\` |"
  echo "| Commits | ${total} individual / ${merges} merge commits |"
  echo "| Components | (작성 시 채움 — Backend, Frontend, Infra 중 실제 변경된 영역만) |"
  echo "| Verification | (작성 시 채움 — 결과 요약만, 명령어 노출 금지) |"
  echo
  echo "(노트 본문 한두 문장 — 무엇을 통합했고 어디까지 push되었는지)"
  echo

  if [[ -n "${BUCKETS[breaking]}" ]]; then
    echo "---"
    echo
    echo "## 0. Breaking Changes"
    echo
    echo "${BUCKETS[breaking]}"
  fi

  echo "---"
  echo
  echo "## 1. Features"
  echo
  if [[ -n "${BUCKETS[features]}" ]]; then echo "${BUCKETS[features]}"; else echo "(없음)"; echo; fi

  echo "---"
  echo
  echo "## 2. Fixes"
  echo
  if [[ -n "${BUCKETS[fixes]}" ]]; then echo "${BUCKETS[fixes]}"; else echo "(없음)"; echo; fi

  if [[ -n "${BUCKETS[ui]}" ]]; then
    echo "---"
    echo
    echo "## 3. User-Interface Refinements"
    echo
    echo "${BUCKETS[ui]}"
  fi

  echo "---"
  echo
  echo "## 4. Infrastructure and Refactors"
  echo
  if [[ -n "${BUCKETS[infra]}" ]]; then echo "${BUCKETS[infra]}"; else echo "(없음)"; echo; fi

  if [[ -n "${BUCKETS[docs]}" ]]; then
    echo "### Documentation"
    echo
    echo "${BUCKETS[docs]}"
  fi

  if [[ -n "${BUCKETS[tests]}" ]]; then
    echo "### Tests"
    echo
    echo "${BUCKETS[tests]}"
  fi

  if [[ -n "${BUCKETS[other]}" ]]; then
    echo "### Other"
    echo
    echo "${BUCKETS[other]}"
  fi

  echo "---"
  echo
  echo "## 5. Verification Summary"
  echo
  echo "(작성 시 채움 — 결과 요약 한 단락 + E2E 시나리오 표)"
  echo
  echo "---"
  echo
  echo "## 6. Known Limitations and Follow-ups"
  echo
  echo "(작성 시 채움 — 후속 작업 bullet list)"
  echo
} > "$OUTPUT_FILE"

echo "Draft written: $OUTPUT_FILE"
echo "  Range: ${PREV_REF}..HEAD"
echo "  Commits: ${total} individual / ${merges} merge"
echo
echo "Next: polish the draft (release-notes skill rules), then run scripts/release.sh ${NEW_VERSION}"
