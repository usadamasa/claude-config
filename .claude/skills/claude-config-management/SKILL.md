---
name: claude-config-management
description: "Claude Code設定(リポジトリルート)の構成管理ガイド。ファイルレベルsymlinkによる設定管理、管理対象の追加・削除、Taskfileタスクの実行方法を提供する。「設定ファイルを追加して」「新しいスキルを追加して」「symlinkの状態を確認して」「Claude設定を変更して」のようにClaude Code設定の構成変更を行うときに使用する。"
---

# Claude Code 構成管理

## アーキテクチャ

`~/.claude` は実ディレクトリ。リポジトリの `dotclaude/` 配下の管理対象ファイルだけを個別に symlink する。
ランタイムファイル(cache, debug, history 等)はリポジトリに含まれない。

```
~/.claude/                       (実ディレクトリ)
├── CLAUDE.md               -> <repo>/dotclaude/CLAUDE-global.md
├── settings.json           -> <repo>/dotclaude/settings.json
├── env.sh                  -> <repo>/dotclaude/env.sh
├── hooks/                  -> <repo>/dotclaude/hooks/
├── skills/
│   ├── <skill-name>/       -> <repo>/dotclaude/skills/<skill-name>/  (symlink)
│   └── <plugin-skills>/       (実ディレクトリ、管理外)
├── cache/                     (ランタイム、管理外)
├── projects/                  (ランタイム、管理外)
└── ...
```

**管理方針**: `dotclaude/` 配下のファイル・ディレクトリが symlink 対象｡リポジトリルートの他のファイルは管理外｡

## タスク

```sh
task setup    # symlink セットアップ (マイグレーション含む)
task status   # 状態確認
task clean    # symlink 削除
```

タスク定義: `Taskfile.yml` (リポジトリルート)

## 管理対象の追加手順

### トップレベルファイルを追加

1. `dotclaude/` に `<filename>` を配置
2. `Taskfile.yml` の `setup` タスクの `for file in dotclaude/settings.json dotclaude/env.sh` に追加
3. `task setup` を実行

### トップレベルディレクトリを追加

1. `dotclaude/` に `<dirname>/` を配置
2. `Taskfile.yml` の `setup` タスクの `for dir in dotclaude/hooks` に追加
3. `task setup` を実行

### グローバルスキルを追加 (git管理対象)

1. `dotclaude/skills/<skill-name>/SKILL.md` を作成
2. `task setup` を実行 (`dotclaude/skills/*/` を自動検出、Taskfile変更不要)

### プロジェクトスコープのスキルを追加

`.claude/skills/<skill-name>/SKILL.md` を作成。symlink 不要。

## worktree 環境チェック

worktree 環境では symlink 先がメインリポジトリを指すため、`~/.claude/` 配下の symlink を直接編集してはいけない。
worktree 判定を行い、`$(pwd)/<file>` を使用すること。

```bash
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON=$(git rev-parse --git-common-dir)
# GIT_DIR != GIT_COMMON → worktree 環境
# GIT_DIR == GIT_COMMON → 通常リポジトリ
```

詳細は CLAUDE.md の「worktree 環境でのファイルパス解決」を参照。

## settings.json 変更ワークフロー

### コミット前の正規化

`settings.json` を変更したら、コミット前に必ず正規化を実行する:

```sh
task settings:normalize  # 正規化 (ソート + ランタイムフィールド除去)
task settings:check      # 正規化済みか検証 (CI と同じチェック)
```

### フィールド分類

| フィールド | 分類 | 扱い |
|---|---|---|
| `permissions.allow/deny/ask` | Intentional | バージョン管理、ソート正規化 |
| `model` | Intentional (pinned) | バージョン管理、CI で値を検証 |
| `hooks` | Intentional | バージョン管理、変更不要 |
| `statusLine` | Intentional | バージョン管理 |
| `sandbox` (含 `allowedDomains`) | Intentional | バージョン管理、ソート正規化 |
| `enabledPlugins` | Intentional | バージョン管理、キー順ソート |
| `language` | Intentional | バージョン管理 |
| `alwaysThinkingEnabled` | Intentional | バージョン管理 |
| `autoMemoryEnabled` | Intentional | バージョン管理 |
| `effortLevel` | Runtime noise | コミット前に除去 |
| `teammateMode` | Runtime noise | コミット前に除去 |

### 自動付与パーミッションの扱い

- Claude Code がセッション中に自動付与したパーミッションは `task settings:normalize` でソートされる
- 意図的な追加: PR で理由を記載してコミット
- 意図しない追加: `git checkout -- dotclaude/settings.json` で revert

### diff レビューチェックリスト

settings.json を含む PR では以下を確認:
1. `effortLevel`, `teammateMode` が含まれていないこと (ランタイムノイズ)
2. `model` の値が意図的な変更か
3. `enabledPlugins` のトグルが意図的か
4. パーミッション配列の変更が意図的か (自動付与 vs 手動追加)

## 注意事項

- `ln -sfn` でディレクトリ先に既存ディレクトリがあるとネスト symlink が発生する。
  `~/.claude` が実ディレクトリであることを前提としている
- `task setup` はべき等 (何度実行しても安全)
