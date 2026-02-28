---
name: improve-config-in-docker
description: "claude-config設定をDockerコンテナ内で改善・検証するスキル。worktreeで設定を変更し、コンテナ内のClaude Codeで挙動を確認する改善ループを回す。マージ前の設定変更をトライアル&エラーで即座に検証できる。「Docker検証」「設定をDockerで試して」「Docker改善ループ」等のトリガーに対応。"
---

# improve-config-in-docker

claude-config設定の変更をDockerコンテナで即座に検証し、改善ループを回すスキル。

## トリガー条件

- 「Docker検証」「設定をDockerで試して」「Docker改善ループ」
- 「コンテナで設定確認」「verify in docker」
- settings.json, env.sh, hooks, skills の変更後に検証したいとき

## Context

- Current branch: !`git branch --show-current`
- Git dir: !`git rev-parse --git-dir`
- Git common dir: !`git rev-parse --git-common-dir`
- Docker image: !`docker image inspect claude-config-verify --format '{{.ID}}' 2>/dev/null || echo "not built"`
- Changed files: !`git diff --name-only HEAD 2>/dev/null | grep -E '^dotclaude/' || echo "no dotclaude changes"`

## 前提条件

- Docker Desktop が起動していること
- `docker/verify.sh` がリポジトリに存在すること

## Worktree環境での注意

claude-configはworktreeで開発することが多い。worktree環境では:

- `docker/verify.sh` の `REPO_ROOT` はworktreeのルートを指すので、dotclaude/ 配下の変更はそのまま反映される
- symlinkで管理された設定ファイル(~/.claude/settings.json等)はメインリポジトリを指すため、worktreeの変更は直接反映されない
- **だからこそDocker検証が有効**: worktreeの設定をコンテナに展開して即座に動作確認できる

## ワークフロー

### Step 1: 変更内容の把握

```bash
git diff --stat HEAD
git diff dotclaude/
```

何が変更されたかを把握し、検証の焦点を決める。

### Step 2: Dockerコンテナの起動と検証

verify.shはTTY判定を内蔵しているので、ターミナルでもCLI環境でも同じコマンドで使える:

```bash
# bashシェルを起動 (ターミナルから)
./docker/verify.sh

# 特定コマンドを実行 (ターミナル/CLI両対応)
./docker/verify.sh claude --version
./docker/verify.sh sh -c 'jq .hooks ~/.claude/settings.json'
```

### Step 3: 検証項目チェックリスト

設定変更の種類に応じて検証:

#### settings.json 変更時
```bash
jq '.permissions' ~/.claude/settings.json
jq '.hooks' ~/.claude/settings.json
jq '.enabledPlugins' ~/.claude/settings.json  # Docker内では全てfalse
```

#### env.sh 変更時
```bash
source ~/.claude/env.sh && env | grep -E '^(CLAUDE|ANTHROPIC|CLOUD_ML)'
```

#### hooks 変更時
```bash
ls -la ~/.claude/hooks/
bash ~/.claude/hooks/guard-home-dir.sh  # 直接テスト
```

#### skills 変更時
```bash
ls ~/.claude/skills/
cat ~/.claude/skills/SKILL_NAME/SKILL.md
```

#### CLAUDE-global.md 変更時
```bash
cat ~/.claude/CLAUDE.md | head -20
```

### Step 4: 結果判定と次のアクション

- 期待通り → 変更をコミット
- 問題あり → 修正してStep 2に戻る

## 改善ループの終了判断

改善ループは以下の条件を満たしたときに終了する:

1. **改善幅の収束**: Docker内Claude Codeからの改善提案が、前回イテレーションと比較して十分に小さくなったとき
   - 「Critical/セキュリティ」レベルの提案が0件
   - 残る提案が「Nice to have」や「将来検討」レベルのみ
2. **テスト全通過**: `bats tests/` が全件パス
3. **Docker内検証OK**: 変更後の設定がDocker内で正しく展開・動作することを確認済み

### イテレーション記録

各イテレーションで以下を記録し、改善幅の推移を追跡する:

| イテレーション | 提案数 | Critical | Important | Suggestion |
|---------------|--------|----------|-----------|------------|
| N             | X      | X        | X         | X          |
| N+1           | Y      | Y        | Y         | Y          |

Critical == 0 かつ Important が前回比で減少傾向にあればループ終了を検討する。

### ユーザー確認が必要なケース

以下の場合は改善を自動適用せず、ユーザーに方針確認を取ること:

- **セキュリティリスクの増大**: パーミッション大幅緩和、deny リストの縮小、認証情報の新規マウント等
- **コンテキスト使用量の増大**: CLAUDE.md の大幅な追記、スキル定義の肥大化、セッションデータの追加読み込み等
- **環境への破壊的変更**: hooks の動作変更、既存スキルの削除・大幅リファクタ、settings.json の構造変更等
- **設計方針の転換**: Docker構成の根本変更、symlink管理方式の変更等

## イメージ管理

```bash
./docker/verify.sh --rebuild  # 強制再ビルド
task docker:clean                # 全リソース削除
```

## パーミッション

- `claude` コマンド実行時、entrypointが自動的に `--dangerously-skip-permissions` を付与する
- Docker内はホストから隔離されているため安全
- `./docker/verify.sh claude -p "設定を確認して"` → パーミッション確認なしで即実行

## 制限事項

- MCP plugins はコンテナ内で動作しないため全て無効化
- statusLine はコンテナ内で不要なため削除
- gcloud ADC はホストの `~/.config/gcloud` が存在する場合のみマウント
