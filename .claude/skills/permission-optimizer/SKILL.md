---
name: permission-optimizer
description: >-
  settings.jsonのBash/Read/Write/Editパーミッションを最適化するスキル。
  セッションJSONLから過去30日間のツール使用状況を分析し、
  安全性評価に基づいてパーミッションの追加/削除を提案する。
  (1) パーミッションの棚卸し (2) 新規パーミッションの追加提案
  (3) 未使用パーミッションの削除提案 (4) ベアエントリ警告 に使用する。
---

# Permission Optimizer

settings.jsonの`permissions.allow`/`permissions.deny`/`permissions.ask`に登録されているBash/Read/Write/Editパーミッションを管理するワークフロー。

## ワークフロー

### 0. worktree 環境チェック

CLAUDE.md の「worktree 環境でのファイルパス解決」を参照し、worktree 判定を行う。
worktree 環境の場合、以降のステップで settings.json のパスを `$(pwd)/dotclaude/settings.json` に読み替えること。

### 1. 分析の実行(エージェント経由)

**推奨: permission-auditor エージェントを使用する。**

コンテキスト圧迫を防ぐため、Task ツールで `permission-auditor` エージェントを起動する。エージェントが CLI を実行し、フル JSON を内部で処理して、コンパクトなサマリのみを返す。

```
Task(subagent_type="permission-auditor", prompt="パーミッション分析を実行してサマリを返してください")
```

**代替: CLI 直接実行(サマリモード)**

エージェントを使わない場合は、CLI のサマリ出力を使う。**`--settings` は必ず指定すること。**

まず settings.json のパスを解決する:
```bash
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON=$(git rev-parse --git-common-dir)
if [ "$GIT_DIR" != "$GIT_COMMON" ]; then
  SETTINGS_PATH="$(pwd)/dotclaude/settings.json"
else
  GIT_ROOT=$(git rev-parse --show-toplevel)
  SETTINGS_PATH="$GIT_ROOT/dotclaude/settings.json"
fi
```

CLI を実行:
```bash
go run ./cmd/analyze-permissions --days 30 --settings "$SETTINGS_PATH" --projects-dir ~/.claude/projects
```

フル JSON が必要な場合は `--format json --output /tmp/report.json` を追加する。

CLIオプション:
- `--days N`: 集計期間を指定(デフォルト: 30日)
- `--settings PATH`: settings.jsonのパスを指定(**必須**: 省略時は自動解決を試みるが、明示指定を推奨)
- `--format summary|json`: 出力形式(デフォルト: summary)
- `--output PATH`: フル JSON をファイルに書き出し(summary と併用可)

### 2. サマリの確認

サマリ出力は以下のセクションを含む:

- **[ADD to allow]**: 追加推奨パーミッション(safe カテゴリで未登録のもの、上位10件)
- **[REVIEW]**: 要確認パーミッション(review/ask カテゴリ、deny バイパスリスク)
- **[UNUSED]**: 未使用パーミッション(リストにあるが使用されていない)
- **[WARNINGS]**: ベアエントリ警告、deny バイパスリスク警告、allow が deny を包含する警告
- **Summary**: 現在の件数と推奨変更数

### 3. ユーザー確認ポイント

settings.json を更新する前に、以下の方針をユーザーに確認する:

1. **curl/wget の取り扱い**: deny に維持するか? (WebFetch 推奨。API テスト等で必要なら project local で allow に)
2. **git remote 操作**: git push / git rebase を ask に維持するか? (安全性とフロー効率のトレードオフ)
3. **破壊操作**: rm -rf の方針 (ask 推奨。deny にすると作業効率が低下)
4. **開発ツールの配置先**: global vs project で適切か? (→ スコープ別ガイドライン参照)

### 4. settings.jsonの更新

ユーザーの承認を得た上で、以下の手順でsettings.jsonを更新する:

1. 承認されたパーミッションのみ適切なリスト(allow/deny/ask)に追加
2. パーミッション形式: `Tool(pattern:*)` (Bash) または `Tool(pattern)` (Read/Write/Edit)
3. 許可エントリはアルファベット順にソート
4. 不要と判断されたエントリは削除
5. ベアエントリが含まれていないことを最終確認

### 5. 変更後の検証

設定変更は新しいセッションから有効になる。以下を確認:
- git コマンド(status/commit/push)が期待通りに動作するか
- deny 対象(curl, ssh 等)が拒否されるか
- 新しいセッションでパーミッションプロンプトの頻度が適切か

## スコープ別配置ガイドライン

| スコープ | 用途 | 例 |
|----------|------|-----|
| global (`~/.claude/settings.json`) | プロジェクト横断の汎用コマンド | git, go, task, make, gh, docker |
| project shared (`.claude/settings.json`) | プロジェクト固有ツール | npm, cargo, プロジェクト固有パス |
| project local (`.claude/settings.local.json`) | 一時的・個人的パーミッション | gcloud, 一時的な curl 許可 |

詳細は `references/recommended-settings.md` を参照。

## パーミッション評価順序の注意

Claude Codeのパーミッション評価は **deny → ask → allow** の順序で行われる。

### ベア(修飾子なし)エントリの禁止

`ask`や`allow`配列にベアの`Bash`や`Read`(修飾子なし)を入れてはいけない。

**誤った設定:**
```json
"ask": ["Bash"],
"allow": ["Bash(git status:*)", "Bash(go test:*)"]
```

この場合、ベアの`Bash`がすべてのBash呼び出しにマッチし、`allow`のコマンド別許可が全て無視される。

**正しい設定:**
```json
"allow": ["Bash(git status:*)", "Bash(go test:*)"],
"ask": ["Bash(git commit:*)", "Bash(git push:*)"]
```

### このスキル実行時の検証

パーミッション管理の更新時に、以下を必ず検証すること:

1. `ask`/`allow`配列にベアの`Bash`、`Read`、`Write`、`Edit`が含まれていないか確認する
2. 含まれている場合はユーザーに警告し、削除を提案する
3. レポートの`bare_entry_warnings`フィールドで自動検出される

## 安全性カテゴリ

### Bash コマンド

| カテゴリ | 説明 | 例 | deny バイパスリスク |
|----------|------|-----|-----|
| safe | 読取系・ビルドツール | git status, go test, task, make, brew list | なし |
| ask | 変更・破壊操作 | git commit, git push, git rebase, rm -rf | なし |
| deny | 外部通信・特権操作 | curl, wget, sudo, ssh, scp, eval | - |
| review (bypass) | Read/Write deny バイパス | cat, head, tail, grep, echo, find, sed, awk | あり |

### deny バイパスリスクとは

Claude Code はツールごとに独立してパーミッションを評価する｡`Read(~/.ssh/**)` を deny していても `Bash(cat:*)` を allow すると `cat ~/.ssh/id_rsa` で deny が回避される｡

| Bash コマンド | バイパス対象 | 代替ツール |
|--------------|-------------|-----------|
| cat, head, tail, grep, awk | Read deny | Read, Grep ツール |
| echo, tee, cp, mv | Write deny | Write ツール |
| find | Read + Write deny (破壊操作含む) | Glob ツール |
| sed | Read + Write deny | Edit ツール |

**推奨**: Claude Code には Read/Grep/Write/Edit/Glob 等の専用ツールがあり、Bash でのファイル操作は原則不要｡deny バイパスリスクのあるコマンドは allow に含めない｡

### Read/Write/Edit パス

| カテゴリ | 説明 | 例 |
|----------|------|-----|
| safe | プロジェクトファイル・設定 | src/**, CLAUDE.md, .claude/** |
| deny | 機密ファイル | ~/.ssh/**, ~/.aws/**, .env, credentials |
| review | 手動確認が必要 | 上記に該当しないパス |

## パーミッション集約の注意事項

複数のスコープ付きエントリを1つに集約する場合、意図せずスコープが拡大しないよう注意する。

**悪い例** (スコープ拡大):
```json
// Before: .claude ディレクトリのみ
"Bash(mkdir -p ~/.claude/**)", "Bash(mkdir .claude/**)"
// After: ファイルシステム全体
"Bash(mkdir:*)"
```

**良い例** (スコープ維持):
```json
// フォーマット修正のみ、スコープは維持
"Bash(mkdir -p:*)", "Bash(mkdir .claude:*)", "Bash(mkdir ~/.claude:*)"
```

## パーミッション形式

- Bash: `Bash(コマンドプレフィックス:*)` 例: `Bash(git status:*)`
- Read: `Read(パスパターン)` 例: `Read(~/.claude/**)`
- Write: `Write(パスパターン)` 例: `Write(src/**)`
- Edit: `Edit(パスパターン)` 例: `Edit(~/.claude/**)`
