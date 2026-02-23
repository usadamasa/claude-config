# CI 統合ガイド (GitHub Actions)

## ツールバージョン管理: aqua

CI でも同じ `aqua.yaml` を使ってツールバージョンを固定する。

### aqua.yaml (プロジェクトルート)

> **注意**: aqua は go-arch-lint の管理に主に使う。
> golangci-lint は GitHub Actions では `golangci/golangci-lint-action` が自前でインストールするため aqua 不要。
> ローカル環境では `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` でも可。

```yaml
---
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/aqua-yaml.json
aqua_version: ">=2.0.0"

registries:
  - type: standard
    ref: v4.227.0  # 定期的に更新する

packages:
  - name: golangci/golangci-lint@v2.9.0   # golangci-lint v2 系を使う (v1 と設定非互換)
  - name: fe3dback/go-arch-lint@v1.14.0
```

---

## GitHub Actions ワークフロー

### 方法 1: aqua でツールを統一管理 (推奨)

```yaml
# .github/workflows/arch-metrics.yml
name: Architecture Metrics

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  setup-tools:
    name: Setup Tools via aqua
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup aqua
        uses: aquaproj/aqua-installer@v3
        with:
          aqua_version: v2.43.0

      - name: Install tools
        run: aqua install

  golangci-lint:
    name: golangci-lint (Testability & Maintainability)
    runs-on: ubuntu-latest
    needs: setup-tools
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Setup aqua
        uses: aquaproj/aqua-installer@v3
        with:
          aqua_version: v2.43.0

      - name: Install tools
        run: aqua install

      - name: Run golangci-lint
        run: golangci-lint run --timeout 5m ./...

  go-arch-lint:
    name: go-arch-lint (Modularity)
    runs-on: ubuntu-latest
    needs: setup-tools
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Setup aqua
        uses: aquaproj/aqua-installer@v3
        with:
          aqua_version: v2.43.0

      - name: Install tools
        run: aqua install

      - name: Check architecture rules
        run: go-arch-lint check ./...
```

### 方法 2: golangci-lint 公式 Action を使う (golangci-lint のみ)

```yaml
# .github/workflows/golangci-lint.yml
name: golangci-lint

on:
  push:
    branches: [main]
  pull_request:

jobs:
  golangci-lint:
    name: lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: false

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v7
        with:
          version: v1.62.2
          # .golangci.yml を自動で読み込む
          args: --timeout 5m
```

---

## PR ゲートとしての活用

### 新規違反のみをチェック (段階的導入)

既存違反が多い場合は `new-from-rev` で PR 差分のみをチェックする:

```yaml
# .golangci.yml
issues:
  new: true
  new-from-rev: origin/main
```

GitHub Actions 側の設定:

```yaml
- name: golangci-lint (new issues only)
  uses: golangci/golangci-lint-action@v7
  with:
    version: v1.62.2
    args: --new-from-rev=origin/main
```

### 違反数レポートをコメントとして投稿

```yaml
- name: Run golangci-lint with JSON output
  run: |
    golangci-lint run --out-format json ./... 2>/dev/null > lint-result.json || true

- name: Post lint summary
  uses: actions/github-script@v7
  with:
    script: |
      const fs = require('fs');
      const result = JSON.parse(fs.readFileSync('lint-result.json', 'utf8'));
      const issues = result.Issues || [];
      const summary = issues.reduce((acc, issue) => {
        acc[issue.FromLinter] = (acc[issue.FromLinter] || 0) + 1;
        return acc;
      }, {});
      const body = Object.entries(summary)
        .map(([linter, count]) => `- ${linter}: ${count} violations`)
        .join('\n');
      github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: `## Architecture Metrics\n\n${body || 'No violations found!'}`
      });
```

---

## Makefile / Taskfile への統合

### Taskfile.yml (go-task)

```yaml
version: '3'

tasks:
  lint:
    desc: "アーキテクチャメトリクスをチェック"
    deps: [lint:golangci, lint:arch]

  lint:golangci:
    desc: "golangci-lint でテスト可能性・保守性を計測"
    cmd: golangci-lint run --timeout 5m ./...

  lint:arch:
    desc: "go-arch-lint でパッケージ依存方向をチェック"
    cmd: go-arch-lint check ./...

  lint:baseline:
    desc: "ベースライン測定 (現状把握)"
    cmd: bash ~/.claude/skills/usadamasa-go-arch-metrics/scripts/baseline.sh ./
```
