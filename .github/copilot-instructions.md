# Copilot Review Instructions

## Repository Context

This repository manages **Claude Code global configuration** files.
A solo-developer personal project; PRs are authored and merged by the same person.

| Component | Language | Purpose |
|-----------|----------|---------|
| `dotclaude/hooks/` | Bash | Claude Code session hooks (symlinked to `~/.claude/hooks/`) |
| `dotclaude/skills/` | Markdown | Claude Code global skill definitions |
| `dotclaude/settings.json` | JSON | Permission and model configuration |
| `cmd/` | Go | CLI tools for permission analysis |
| `.github/workflows/` | YAML | CI (shellcheck, bats, Go tests) |
| `tests/` | Bash (bats) | Integration tests |

## Review Focus

Prioritize these categories when reviewing PRs:

- **Security**: Command injection, unsafe variable expansion in shell scripts, permission design holes in `settings.json`
- **Correctness**: Logic bugs, macOS/Linux portability issues, incorrect path resolution (especially in worktree environments)
- **settings.json quality**: Bare permission entries (e.g. `Bash` without scope), deny-bypass risks, overly broad wildcards
- **Go CLI correctness**: Pattern matching bugs, incorrect JSON parsing, missing edge cases

## Do Not Flag

The following patterns are intentional or covered by other tools. Do not raise review comments for them:

- **Japanese half-width punctuation** (`｡` `､`): Project style convention, used intentionally per coding standards
- **shellcheck disable comments** (`# shellcheck disable=SC...`): Verified in CI; each disable is reviewed and intentional
- **`curl | bash` in `docker/verify.sh`**: Runs inside a disposable container with no persistence; risk is accepted
- **PR scope / bundled changes**: Solo developer workflow; related changes are intentionally bundled in single PRs
- **Minor documentation inconsistencies**: Low priority; do not flag typos or formatting in `.md` files unless they cause confusion
- **Shell quality rules** (quoting, SC2155, `printf` vs `echo`, `set -euo pipefail`): Enforced by shellcheck + pre-commit hooks in CI. Do not duplicate these checks
- **`|| true` error suppression**: When accompanied by a log message or comment explaining intent, this is acceptable
- **Test assertion style**: Assertion granularity is managed by the developer; do not flag `NotEmpty` vs `Equal` choices
