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
OUTPUT_FILE_KO="${OUTPUT_DIR}/v${NEW_VERSION}.ko.md"
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

# Emit drafts. The author replaces bullet lists with polished prose per the
# release-notes skill rules; these files are just the raw material. Two files
# are produced — English main and Korean translation — sharing the same bullets.

emit_draft() {
  local out="$1"          # path
  local lang="$2"         # "en" or "ko"
  local title_suffix=""
  local cross_link=""
  local placeholder_components="" placeholder_verify="" placeholder_intro=""
  local placeholder_features="" placeholder_fixes="" placeholder_infra=""
  local label_features="" label_fixes="" label_ui="" label_infra="" label_docs="" label_tests="" label_other=""
  local label_breaking="" label_verify="" label_known="" verify_body="" known_body=""

  if [[ "$lang" == "en" ]]; then
    title_suffix=""
    cross_link="> [한국어 버전 / Korean version](./v${NEW_VERSION}.ko.md)"
    placeholder_components="(fill in — Backend, Frontend, Infra; whichever actually changed)"
    placeholder_verify="(fill in — outcomes only, do not paste verification shell commands)"
    placeholder_intro="(one or two sentences — what was integrated and how far it was pushed)"
    placeholder_features="(none)"
    placeholder_fixes="(none)"
    placeholder_infra="(none)"
    label_breaking="## 0. Breaking Changes"
    label_features="## 1. Features"
    label_fixes="## 2. Fixes"
    label_ui="## 3. User-Interface Refinements"
    label_infra="## 4. Infrastructure and Refactors"
    label_docs="### Documentation"
    label_tests="### Tests"
    label_other="### Other"
    label_verify="## 5. Verification Summary"
    label_known="## 6. Known Limitations and Follow-ups"
    verify_body="(fill in — one paragraph of outcomes + E2E scenario table)"
    known_body="(fill in — follow-up bullet list)"
  else
    title_suffix=" (한국어)"
    cross_link="> [English version](./v${NEW_VERSION}.md)"
    placeholder_components="(작성 시 채움 — Backend, Frontend, Infra 중 실제 변경된 영역만)"
    placeholder_verify="(작성 시 채움 — 결과 요약만, 명령어 노출 금지)"
    placeholder_intro="(노트 본문 한두 문장 — 무엇을 통합했고 어디까지 push되었는지)"
    placeholder_features="(없음)"
    placeholder_fixes="(없음)"
    placeholder_infra="(없음)"
    label_breaking="## 0. Breaking Changes"
    label_features="## 1. Features"
    label_fixes="## 2. Fixes"
    label_ui="## 3. User-Interface Refinements"
    label_infra="## 4. Infrastructure and Refactors"
    label_docs="### Documentation"
    label_tests="### Tests"
    label_other="### Other"
    label_verify="## 5. Verification Summary"
    label_known="## 6. Known Limitations and Follow-ups"
    verify_body="(작성 시 채움 — 결과 요약 한 단락 + E2E 시나리오 표)"
    known_body="(작성 시 채움 — 후속 작업 bullet list)"
  fi

  {
    echo "# Release Notes — v${NEW_VERSION}${title_suffix}"
    echo
    echo "${cross_link}"
    echo
    echo "| | |"
    echo "|---|---|"
    echo "| Range | \`${PREV_REF}..HEAD\` |"
    echo "| Commits | ${total} individual / ${merges} merge commits |"
    echo "| Components | ${placeholder_components} |"
    echo "| Verification | ${placeholder_verify} |"
    echo
    echo "${placeholder_intro}"
    echo

    if [[ -n "${BUCKETS[breaking]}" ]]; then
      echo "---"
      echo
      echo "${label_breaking}"
      echo
      echo "${BUCKETS[breaking]}"
    fi

    echo "---"
    echo
    echo "${label_features}"
    echo
    if [[ -n "${BUCKETS[features]}" ]]; then echo "${BUCKETS[features]}"; else echo "${placeholder_features}"; echo; fi

    echo "---"
    echo
    echo "${label_fixes}"
    echo
    if [[ -n "${BUCKETS[fixes]}" ]]; then echo "${BUCKETS[fixes]}"; else echo "${placeholder_fixes}"; echo; fi

    if [[ -n "${BUCKETS[ui]}" ]]; then
      echo "---"
      echo
      echo "${label_ui}"
      echo
      echo "${BUCKETS[ui]}"
    fi

    echo "---"
    echo
    echo "${label_infra}"
    echo
    if [[ -n "${BUCKETS[infra]}" ]]; then echo "${BUCKETS[infra]}"; else echo "${placeholder_infra}"; echo; fi

    if [[ -n "${BUCKETS[docs]}" ]]; then
      echo "${label_docs}"
      echo
      echo "${BUCKETS[docs]}"
    fi

    if [[ -n "${BUCKETS[tests]}" ]]; then
      echo "${label_tests}"
      echo
      echo "${BUCKETS[tests]}"
    fi

    if [[ -n "${BUCKETS[other]}" ]]; then
      echo "${label_other}"
      echo
      echo "${BUCKETS[other]}"
    fi

    echo "---"
    echo
    echo "${label_verify}"
    echo
    echo "${verify_body}"
    echo
    echo "---"
    echo
    echo "${label_known}"
    echo
    echo "${known_body}"
    echo
  } > "$out"
}

emit_draft "$OUTPUT_FILE" en
emit_draft "$OUTPUT_FILE_KO" ko

echo "Drafts written:"
echo "  English: $OUTPUT_FILE"
echo "  Korean:  $OUTPUT_FILE_KO"
echo "  Range:   ${PREV_REF}..HEAD"
echo "  Commits: ${total} individual / ${merges} merge"
echo
echo "Next: polish both drafts (release-notes skill rules — English is the main"
echo "      file that becomes the GitHub Release body), then run scripts/release.sh ${NEW_VERSION}"
