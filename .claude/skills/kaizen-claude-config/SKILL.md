---
name: kaizen-claude-config
description: "claude-configのグローバル設定(dotclaude/)をDockerコンテナ内で自己評価・改善するスキル。worktreeの設定変更をコンテナに展開し、Docker内のClaude Codeに設定を評価させて改善提案を得る。「Docker検証」「設定をDockerで試して」「verify in docker」「設定改善」「kaizen」等のトリガーに対応。"
---

# kaizen-claude-config

## 検証スコープ

**グローバル設定 (dotclaude/ → ~/.claude/)** のみが対象。

| 対象 | 説明 |
|---|---|
| dotclaude/CLAUDE-global.md | グローバル CLAUDE.md |
| dotclaude/settings.json | パーミッション・フック・プラグイン設定 |
| dotclaude/env.sh | 環境変数設定 |
| dotclaude/hooks/ | セッションフック |
| dotclaude/bin/ | フック依存バイナリ |
| dotclaude/skills/*/ | グローバルスキル |

**対象外**: プロジェクトスコープの設定 (.claude/settings.local.json, .claude/agents/, プロジェクト CLAUDE.md)

## Context

- Current branch: !`git branch --show-current`
- Git dir: !`git rev-parse --git-dir`
- Git common dir: !`git rev-parse --git-common-dir`
- Docker image: !`docker image inspect claude-config-verify --format '{{.ID}}' 2>/dev/null || echo "not built"`
- Changed files: !`git diff --name-only HEAD 2>/dev/null | grep -E '^dotclaude/' || echo "no dotclaude changes"`

## 前提条件

- Docker Desktop が起動していること
- `docker/verify.sh` がリポジトリに存在すること

## ワークフロー

### Step 1: 変更内容の把握

```bash
git diff --stat HEAD
git diff dotclaude/
```

何が変更されたかを把握し、検証の焦点を決める。

### Step 2: Docker 検証の実行

**config-evaluator エージェントに委譲する。** メインコンテキストで verify.sh を直接実行しない。

config-evaluator エージェントを Task tool で起動する:
- `subagent_type: "general-purpose"`
- prompt に `.claude/agents/config-evaluator.md` の参照と検証指示を含める

### Step 3: 結果判定と次のアクション

エージェントから返されたサマリに基づいて判断:

- **全 OK** → 変更をコミット
- **問題あり** → 修正して Step 2 に戻る

### ユーザー確認が必要なケース

以下の場合は改善を自動適用せず、ユーザーに方針確認を取ること:

- **セキュリティリスクの増大**: パーミッション大幅緩和、deny リストの縮小
- **コンテキスト使用量の増大**: CLAUDE.md の大幅な追記、スキル定義の肥大化
- **環境への破壊的変更**: hooks の動作変更、既存スキルの削除・大幅リファクタ
- **設計方針の転換**: Docker 構成の根本変更、symlink 管理方式の変更

## Worktree 環境での注意

worktree 環境では symlink で管理された設定ファイル (~/.claude/settings.json 等) はメインリポジトリを指すため、worktree の変更は直接反映されない。
`docker/verify.sh` は worktree ルートの dotclaude/ を Docker に展開するため、worktree の設定を即座に検証できる。

## イメージ管理

```bash
./docker/verify.sh --rebuild  # 強制再ビルド
task docker:clean              # 全リソース削除
```

## 制限事項

- gcloud ADC はホストの `~/.config/gcloud` が存在する場合のみマウント
- `claude` コマンド実行時、entrypoint が自動的に `--dangerously-skip-permissions` を付与
