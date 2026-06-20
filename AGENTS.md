# AGENTS.md

## Purpose

These instructions define the required working rules for Codex in this repository. Follow them when reading, planning, editing, testing, reviewing, and committing changes.

## Engineering principles

- Follow Clean Code, KISS, DRY, and SOLID.
- Prefer simple, explicit, maintainable solutions over clever or over-engineered ones.
- Preserve the existing Go project structure and code style unless there is a clear technical reason to improve it.
- Keep user-facing behavior documented in `README.md`; keep Codex workflow, architecture notes, and maintainer reminders in `AGENTS.md` or dedicated developer docs.
- Do not duplicate logic. Extract shared functions or helpers when similar behavior appears in more than one place.

## PocketBook server standards

- Keep product-facing naming aligned with `pb-ftp` for the PocketBook launcher/server and `eBookSender` only when referring to the Android client.
- Treat `/version`, `/rescan`, and `/update` as public local HTTP API contracts. Keep them backward compatible unless the task explicitly requires a breaking change.
- Preserve safe self-update behavior: validate staged paths, reject path traversal and symlinks, verify SHA-256, prevent downgrades by `versionCode`, and keep rollback behavior for launcher replacement.
- Avoid blocking UI or network startup paths longer than needed on the PocketBook device.
- Do not log or expose secrets. The FTP server is anonymous by design; do not add credential handling without explicit product direction.

## Changelog and localization

- Keep `CHANGELOG.md` and `CHANGELOG.ru.md` updated for user-facing changes, release notes, or behavior changes.
- `CHANGELOG.md` is the English fallback changelog. `CHANGELOG.ru.md` is the Russian bundled changelog.
- The release workflow publishes localized changelog files under `updates/changelog/` and exposes them in `updates/latest.json` through `changelogUrls`.
- When adding or changing user-facing release-note content, update both changelog files in the same change.
- At task completion, state whether changelog updates were needed.

## Working process

- Start by reading relevant files and existing analogues before editing.
- Use `rg` and `rg --files` for search.
- For large or ambiguous tasks, create a concise professional plan first and wait for approval before implementation.
- For small, obvious fixes, proceed directly on a task-appropriate branch while keeping scope tight.
- At task completion, decide whether `AGENTS.md` needs updates. Update it yourself when workflow, architecture, important paths, shared patterns, or verification commands changed; otherwise state that no guidance update was needed.

## Git workflow

- Check `git status --short --branch` before making changes.
- Do all work on a task-appropriate branch, not directly on `main`, unless the user explicitly asks otherwise.
- For new features, create a dedicated branch named `feature/<short-name>`.
- For bug fixes, create a dedicated branch named `bugfix/<short-name>`.
- For refactoring and small maintenance tasks, use `refactoring` or `refactor/<short-name>` when a separate branch is useful.
- Never overwrite, reset, or revert user changes unless explicitly requested.
- Commit every completed code change that modifies more than five lines, unless the user explicitly asks not to commit.
- Use professional commit messages: write a concise imperative subject that names the actual change, keep it specific enough to stand alone in history, and avoid vague subjects such as `fix`, `update`, or `changes`.
- Keep commits logically grouped. Stage and commit only the files that belong to the current task.

## Verification

- Run the smallest relevant verification first, then broaden when the blast radius is larger.
- For Go code changes, run:

  ```sh
  go test ./...
  ```

- In GitHub Actions, exclude `cmd/app` from the plain Ubuntu unit-test step because it links against PocketBook `inkview`; the workflow still builds `cmd/app` inside the PocketBook SDK Docker image.

- For release workflow or packaging changes, inspect `.github/workflows/ci-cd.yml` and verify the generated manifest shape against the Android client model when possible.
- If verification cannot be run, state the exact reason and residual risk.
