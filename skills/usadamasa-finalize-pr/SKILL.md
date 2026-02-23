---
name: finalize-pr
description: >
  PRレビュー実行→自動マージ→worktreeクリーンアップ→セッション引き継ぎの一連のワークフロー。
  PRがマージ可(Critical Issue 0件)と判断したら、ユーザーに問い合わせずにマージし、
  次のセッションへの引き継ぎまで自動で行う。
  「PRをマージして」「レビューしてマージ」「finalize」「仕上げて」等の手動トリガーに対応。
  PR作成後やレビュー後にプロアクティブに呼び出すことも推奨。
---

# finalize-pr

PRレビューからマージ、クリーンアップ、セッション引き継ぎまでを一気通貫で実行するスキル。

## トリガー条件

- **手動**: 「PRをマージして」「レビューしてマージ」「finalize」「仕上げて」等
- **プロアクティブ**: PR作成後、またはレビュー完了後に自動的に呼び出す

## Context

- Current branch: !`git branch --show-current`
- Git status: !`git status --short`
- PR status: !`gh pr view --json number,title,state,mergeable,mergeStateStatus,statusCheckRollup 2>/dev/null || echo "No PR found"`

## ワークフロー

### Step 1: PR の特定

現在のブランチに紐づく PR を取得する。

```bash
gh pr view --json number,title,url,headRefName,baseRefName,state,mergeable,mergeStateStatus
```

PR が見つからない場合はエラーとして終了する。
PR が既にマージ済みの場合は Step 6 (クリーンアップ) へスキップ。

### Step 2: CI 確認

プロジェクトの CI コマンドを実行する。

CI コマンドの決定順序:
1. CLAUDE.md に `task ci` の記載がある → `task ci` を実行
2. Taskfile.yml に `ci` タスクがある → `task ci` を実行
3. Makefile に `ci` ターゲットがある → `make ci` を実行
4. いずれもない → CI はスキップ (GitHub Actions に委譲)

CI が失敗した場合は即座に停止し、失敗内容を報告する。
**CI が通るまでレビューには進まない。**

### Step 3: PR レビュー実行

`review-pr` のワークフローに従い、以下のレビューエージェントを**並列**で起動する。

起動するエージェント (変更ファイルに応じて選択):
- **code-reviewer**: 常に起動
- **pr-test-analyzer**: テストファイルが変更されている場合
- **comment-analyzer**: コメント/ドキュメントが追加されている場合
- **silent-failure-hunter**: エラーハンドリングが変更されている場合
- **type-design-analyzer**: 型が追加/変更されている場合

各エージェントの起動には Task ツールの `pr-review-toolkit:*` サブエージェントを使用する。

### Step 4: 結果集約・マージ判定

レビュー結果を以下のカテゴリに分類する:

| カテゴリ | マージへの影響 |
|---------|--------------|
| Critical Issues | **マージをブロック** |
| Important Issues | レポートのみ (ブロックしない) |
| Suggestions | レポートのみ (ブロックしない) |

#### マージ判定

- **Critical Issue == 0**: マージ可 → Step 5 へ進む
- **Critical Issue >= 1**: マージ不可 → 問題を報告して**停止**

マージ不可の場合、以下の形式で報告する:

```markdown
## Merge Blocked

Critical Issues が検出されたためマージできません。

### Critical Issues (X件)
- [agent-name]: Issue description [file:line]

### Important Issues (X件)
- [agent-name]: Issue description [file:line]

### Suggestions (X件)
- [agent-name]: Suggestion [file:line]

修正後、再度 `/finalize-pr` を実行してください。
```

### Step 5: PR マージ

マージ方法の決定:
1. CLAUDE.md / MEMORY.md にマージ方法の記載がある → その方法に従う
2. 記載がない → `gh pr merge --merge` (merge commit)

merge queue が有効な場合:
```bash
gh pr merge --merge --auto
```

merge queue が無効な場合:
```bash
gh pr merge --merge
```

マージ成功時、Important Issues / Suggestions があればサマリーを表示する。

### Step 6: worktree クリーンアップ

worktree 環境かどうかを判定する:

```bash
cat .git
```

- **`.git` がファイル** (内容が `gitdir: ...`) → worktree 環境
- **`.git` がディレクトリ** → 通常環境 (クリーンアップ不要、Step 7 へ)

worktree 環境の場合:

1. worktree パスを記録:
   ```bash
   WORKTREE_PATH=$(pwd)
   ```

2. main worktree (bare repo) のパスを取得:
   ```bash
   MAIN_WORKTREE=$(git worktree list --porcelain | head -1 | sed 's/worktree //')
   ```

3. ユーザーに worktree 削除を案内する:
   ```
   worktree を削除するには、別のターミナルで以下を実行してください:
   cd $MAIN_WORKTREE && git worktree remove $WORKTREE_PATH
   ```

**注意**: リモートブランチは削除しない (GitHub が PR マージ時に自動削除する設定を想定)。

### Step 7: セッション引き継ぎ

`session-handoff` スキルを呼び出してセッションを引き継ぐ。

引き継ぎ内容に以下を含める:
- マージした PR の番号と URL
- レビューで検出された Important Issues / Suggestions (あれば)
- 次のタスクへの示唆 (あれば)

## 注意事項

- CI 失敗時はレビューに進まない (CI 優先ポリシー)
- Critical Issue が 1件でもあればマージしない
- Important Issues / Suggestions はマージをブロックしないがレポートする
- worktree のクリーンアップは案内のみ (現在のセッションからは削除できない)
- マージ方法はプロジェクトの CLAUDE.md / MEMORY.md の記載に従う
- このスキルの実行中にエラーが発生した場合は、その時点で停止して報告する
