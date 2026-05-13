# NAS Project Rules

Project-specific rules. **Supplement** (not replace) `~/.claude/CLAUDE.md` global rules.

# ═══════════════════════════════════════════════════

# 🔴 PROJECT HARD RULES (NON-NEGOTIABLE)

# ═══════════════════════════════════════════════════

These rules exist because of specific failures on this project. Every rule cites the date and incident that motivated it. Treat as inviolable.

---

## RULE 1: NO FRONTEND WORK IS COMPLETE WITHOUT PLAYWRIGHT VERIFICATION

**Trigger**: ANY change to:
- `frontend/src/**/*.svelte`
- `frontend/src/**/*.ts`
- `frontend/src/**/*.css`
- `frontend/src/app.css`
- `frontend/svelte.config.js` / `vite.config.ts`

**Forbidden actions before Playwright verification completes**:
- Writing `DONE`, `✅`, `complete`, `ready to merge`, `all checks pass`, `production-ready` in any user-facing report
- `git merge` into `main`
- `git push origin main`
- `gh pr create`
- Marking a TodoWrite item `completed` for a frontend task

**Required workflow** (every bullet is mandatory, no shortcuts):

1. Start dev server in background:
   ```bash
   cd frontend && npm run dev
   ```
   (Use `run_in_background: true` in Bash tool. Wait for `Local: http://localhost:5173` output via Monitor or sleep then read.)

2. Navigate via Playwright MCP:
   ```
   mcp__playwright__browser_navigate("http://localhost:5173")
   ```

3. Authenticate. Use existing local-login credentials (anxi77 dev account) or test session. If the change is auth-related, exercise BOTH the affected path AND the unaffected paths to confirm no regression.

4. Walk through EVERY interactive control changed in the task. For each:
   - Click / type / hover / press the relevant input
   - Verify the resulting state visually
   - Capture screenshot via `mcp__playwright__browser_take_screenshot`
   - Save to `claudedocs/<task-slug>/screenshots/<step-name>.png`

5. For each row of the plan's spec walkthrough table (if any), record `✅` or `❌` with the screenshot reference.

6. Append a verification block to the md trail:
   ```markdown
   ## [MODE: VERIFY] Playwright walkthrough — <timestamp>
   - Spec row 1 (open video file): ✅ screenshots/01-video-loaded.png
   - Spec row 2 (click play): ✅ screenshots/02-playing.png
   - ...
   - Spec row N (keyboard ←): ❌ FAILED — actual: seeks -10s, expected: -5s
   ```

7. If any row is ❌: fix it (separate commit on the feature branch or a follow-up patch PR). Do NOT proceed to merge/push.

**Why this rule exists**:
On **2026-05-13** the `feat/media-viewers` branch (15 commits implementing pro media player UI) was merged into `main` with **zero browser verification**. `npm run check`, `npm run build`, `go test`, and `/code-review` all passed. The controller declared `✅ DONE` and pushed. The user then asked "did you actually Playwright test?" — exposing that automated checks verify code correctness, NOT feature correctness. UI bugs that pass `svelte-check` (broken event bindings, CSS specificity issues, runtime lucide-svelte component rendering failures, keyboard event conflicts with parent listeners) are invisible until clicked in a real browser. This rule prevents that from happening again.

**Failure consequence**: If you skip this rule and the user later finds a broken control, you must (1) acknowledge the skip explicitly, (2) run Playwright verification against current `origin/main`, (3) fix every finding with a follow-up patch PR. Do not minimize the gap as "edge case I missed".

---

## RULE 2: SPEC WALKTHROUGH TABLE IS A HARD GATE

**Trigger**: Any plan file (`Docs/superpowers/plans/*.md`) that contains a spec walkthrough table or enumerable acceptance criteria (e.g., "Spec table: | # | Action | Expected |").

**Action**: Every row must be:
- Verified individually (no batch claims like "all 27 rows pass")
- Recorded with `✅` or `❌` + concrete evidence (screenshot path, log line, command output) in the md trail
- Compared against the plan's literal expected behavior — not a paraphrase, not "looks roughly right"

**Forbidden**: `Spec walkthrough: TBD`, `Spec rows: assumed correct`, `Spec table: skipped (covered by build)`, or any equivalent hedge that drops accountability.

---

## RULE 3: USER-REQUESTED REVIEW STAGES ARE NON-NEGOTIABLE

**Trigger**: User explicitly requests N-stage review (Korean: "양단 review", "이중 review", "리뷰 두 단계"; English: "two-stage review", "spec + code quality reviewer", "double review").

**Action**: Every requested stage MUST run as a **separate subagent dispatch** per task. Each stage's findings must be recorded in the md trail.

**Forbidden**:
- Substituting controller-side inline diff inspection for a fresh spec-compliance subagent
- Substituting the implementer's own `/code-review` for an independent code-quality subagent
- Justifying a skip with "the task is mechanical", "the plan is verbatim", "the file is small", "efficiency", "token cost", or "duration"

**Why this rule exists**:
On **2026-05-13** the controller dropped spec-compliance reviewer subagent dispatches after Task 1 "for efficiency". On Task 11 the implementer renamed `onLoadedMetadata` → `onMeta` and `onVolumeChange` → `onVolume` to satisfy the 12-char rule, **breaking cross-component naming consistency** with VideoPlayer. The drift was caught only by the final cross-cutting review (one full stage too late) and required a separate fix commit. Independent reviewers catch drift that the controller's familiarity bias misses.

---

## RULE 4: SUBAGENT COMMIT BOUNDARY

**Trigger**: Any subagent dispatch (Agent tool) doing implementation in this repo.

**Action**: Implementer subagents MUST NOT run `git commit`. Subagent prompts must include this clause verbatim:

```
WORKFLOW: DO NOT commit. After implementation + type check + /code-review pass with 0 ❌ Critical, STOP and report. Controller will commit.
```

The controller commits after:
- Reading the subagent's report
- Verifying file changes match plan spec (inline diff or full Read-back)
- Running any cross-cutting checks (type, build, test)

**Why this rule exists**: On **2026-05-13** implementer subagents repeatedly forgot the `git commit` step at the end of their prompt (Tasks 2, 5, 6, 8) despite escalating prompt emphasis. Controller-side commits are predictable; subagent-side commits are unreliable when the prompt is long or includes a heavy `/code-review` block before the commit step.

---

## RULE 5: GIT FROM PROJECT ROOT, ALWAYS

**Trigger**: Any `git` command via Bash tool.

**Action**: Prefix with explicit project root:
```bash
cd C:/Data/Git/ANXI/nas && git <command>
```
Or use absolute paths in `git add`. Never run `git` immediately after a `cd frontend && npm ...` without re-anchoring.

**Why this rule exists**: On **2026-05-13** Task 2 commit failed with `frontend/frontend/src/...: pathspec did not match` because shell working directory persisted in `frontend/` from the prior `npm run check`. Cost ~30s of confused debugging.

---

## RULE 6: PRE-MERGE / PRE-PR CHECKLIST

**Trigger**: About to do any of:
- `git merge feat/* main` (or fast-forward merge into main)
- `git push origin main`
- `gh pr create`
- Final user-facing "task complete" report for a frontend branch
- Marking the final TodoWrite item as `completed`

**Required (every checkbox must be ticked AND recorded in md trail)**:

- [ ] Every plan task committed
- [ ] `cd frontend && npm run check` → `0 errors, 0 warnings`
- [ ] `cd frontend && npm run build` → succeeds (no Vite errors)
- [ ] If backend touched: `cd backend && go test ./tests/integration/` relevant subset → passes
- [ ] `/code-review` on every changed file → `0 ❌ Critical`
- [ ] **Playwright walkthrough done per Rule 1** (screenshots saved, evidence linked)
- [ ] **Plan's spec walkthrough table: every row `✅`** (Rule 2)
- [ ] Cross-component naming/style consistency verified (if multi-component change)
- [ ] md trail in `claudedocs/<date>-<slug>.md` updated with all verification evidence

If any checkbox is unticked, the merge/push/PR/completion is **BLOCKED**. Tell the user explicitly:
> "I cannot mark this complete because <unchecked item>. Need to <action> first."

Do not silently bypass. Do not promise to do it later. Do it now or escalate.

---

## RULE 7: HONEST STATUS REPORTING

**Trigger**: Any "complete" / "done" / "summary" report to user.

**Action**: Structure status by verification source:

```
## Verified by automation
- npm run check: 0 errors
- npm run build: succeeds
- Go test TestVideoStream: pass
- /code-review on N files: 0 Critical

## Verified by browser (Playwright)
- Spec row 1: ✅ (screenshot link)
- ...

## NOT verified
- <list anything skipped, with reason and TODO>
```

**Forbidden phrases unless backed by enumerated evidence**:
- `All checks pass` → enumerate which checks
- `DONE` / `✅ DONE` / `complete` → must satisfy Rules 1, 2, 6
- `ready to merge` → must satisfy Rule 6
- `production-ready`, `battle-tested`, `robust`, `blazingly fast`, `bulletproof` → forbidden absent evidence; use `MVP`, `untested in production`, `needs validation`

**Why this rule exists**: ~/.claude/RULES.md "Professional Honesty" baseline + the 2026-05-13 incident where `✅ DONE` was declared with zero browser verification.

---

# ═══════════════════════════════════════════════════

# Project Configuration

# ═══════════════════════════════════════════════════

**Stack**:
- Backend: Go + SQLite (`backend/`)
- Frontend: SvelteKit + Svelte 5 runes + Tailwind 4 (Gruvbox-dark) + FiraD2 font (`frontend/`)
- Build: Vite via `@sveltejs/adapter-static`

**Testing reality**:
- Backend: Go test framework, `tests/integration/` is the E2E suite
- Frontend: **NO unit test framework** (per global rule #1 — do not invent fake verifier boilerplate). Frontend feature verification = **Playwright MCP walkthrough** (Rule 1)

**Theme**: Gruvbox-dark mandatory (`~/.claude/projects/C--Data-Git-ANXI-nas/memory/feedback_gruvbox_palette.md`). Yellow accent `#fabd2f` = `--color-accent`. Never propose alt palettes.

**Auth**: Local + Discord + Google OAuth. Token via `auth.svelte.ts` runes store. Kakao deprecated.

**Frontend conventions**:
- Svelte 5 runes (`$state`, `$derived`, `$effect`, `$props`) — no legacy stores in new code
- Tailwind 4 with Gruvbox tokens defined in `frontend/src/app.css` (`bg-bg-overlay`, `text-fg-accent`, `bg-accent`, etc.)
- Lucide icons via `import X from "lucide-svelte/icons/x"`
- Component pattern: state owner + presentational subcomponent (see `VideoPlayer` + `VideoControls`)

**Backend conventions**:
- `http.ServeContent` for ranged media streaming (verified in `TestVideoStream`)
- Token in Authorization header (not query param — see 2026-05-13 TUS fix)

# ═══════════════════════════════════════════════════

# Common Commands

# ═══════════════════════════════════════════════════

```bash
# Dev server (background)
cd frontend && npm run dev

# Type check
cd frontend && npm run check

# Production build
cd frontend && npm run build

# Backend integration tests
cd backend && go test ./tests/integration/ -run <TestName>

# Git (always from project root)
cd C:/Data/Git/ANXI/nas && git checkout -b feat/<name>
cd C:/Data/Git/ANXI/nas && git commit -m "[type] message"
```

**Commit convention** (per ~/.claude/conventions/COMMIT_CONVENTION.md):
- Bracket format: `[feat]`, `[fix]`, `[refactor]`, `[docs]`, `[test]`, `[chore]`
- **NO `Co-Authored-By: Claude` line**
- **NO `Generated with Claude Code` line**
- Plain `-m "..."` only
