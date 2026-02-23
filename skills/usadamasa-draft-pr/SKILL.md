---
name: draft-pr
description: コミットをfixupして1つに集約し、Draft PRを作成するワークフロー。差分からコミットメッセージを自動生成します。「PRを作って」「Draft PRを作成」「コミットをまとめてPR」「push してPR」のように依頼されたときに使用する。
---

# draft-pr

現在のブランチのコミットを1つに集約し、GitHub MCP ツールで Draft PR を作成・更新する。

## ワークフロー概要

1. リポジトリ情報取得 (owner/repo)
2. 親ブランチの検出 (tracking → merge-base距離 → デフォルト)
3. ベースブランチの決定 (SQUASH_BASE / PR_BASE)
4. コミット状況を確認し準備
5. 差分からコミットメッセージを自動生成
6. force push
7. GitHub MCP で Draft PR を作成 or 更新

## Context

- Current branch: !`git branch --show-current`
- Git status: !`git status --short`
- Recent commits: !`git log --oneline -5`

## 手順

### 1. リポジトリ情報の取得

owner と repo を取得（MCP ツールで必須）:

```bash
git remote get-url origin
```

出力例から owner/repo をパース:
- `git@github.com:OWNER/REPO.git` → OWNER, REPO
- `https://github.com/OWNER/REPO.git` → OWNER, REPO

`.git` サフィックスがある場合は除去。

以降、取得した OWNER と REPO を各 MCP ツール呼び出しに使用する。

### 2. 親ブランチの検出

以下の3段階フォールバックで親ブランチ `$PARENT` を検出する。

#### 優先度1: tracking config

```bash
CURRENT=$(git branch --show-current)
PARENT=$(git config --get branch.$CURRENT.merge 2>/dev/null | sed 's|refs/heads/||')
```

`git checkout -b child --track parent` で作成された場合に有効。

#### 優先度2: merge-base distance 比較

tracking が未設定の場合、全リモートブランチとの距離を比較する。

```bash
# 全リモートブランチを取得（現在のブランチとHEADを除外）
# 各ブランチについて merge-base との距離を計算
# git rev-list --count $(git merge-base HEAD origin/<candidate>)..HEAD
# 最小距離のブランチを親とする
```

例: main(距離5) vs feature/parent(距離2) → feature/parent を選択。

#### 優先度3: デフォルトブランチにフォールバック

```bash
git remote show origin | grep 'HEAD branch' | sed 's/.*: //'
```

検出後、AskUserQuestion でユーザーに `$PARENT` が正しいか確認を取る。

### 3. ベースブランチの決定

検出した親ブランチについて以下の2つの変数を決定する:

- `$SQUASH_BASE`: コミット集約の基準点（merge-base のコミットハッシュ）
- `$PR_BASE`: PRのベースブランチ名

```bash
# リモートの最新を取得
git fetch origin $PARENT
```

```bash
# 親ブランチがリモートに存在するか確認
git ls-remote --heads origin "$PARENT" | grep -q "$PARENT"
```

- **存在する** → `$PR_BASE = $PARENT`
- **存在しない** → `$PR_BASE = デフォルトブランチ`（フォールバック、ユーザーに通知）

`$SQUASH_BASE` はリモート存在に関わらず常に merge-base を使用:

```bash
SQUASH_BASE=$(git merge-base HEAD "origin/$PARENT")
```

### 4. 前提条件の確認

- 現在のブランチが `$PR_BASE` と異なることを確認
  - **同じ場合**: AskUserQuestion で新しいブランチ名を確認し `git checkout -b <name>` で作成
- コミット数を確認: `git rev-list --count $SQUASH_BASE..HEAD`
  - **1以上** → ケースA（rebase fixup）
  - **0** → `git status --porcelain` を確認
    - 変更あり → ケースB（新規コミット）
    - 変更なし → エラー: 変更がないためPR作成不可

### 5. コミット準備

#### ケースA: 差分コミットあり（rebase fixup）

1. 未コミット変更があれば先にコミット:
   ```bash
   git add -A && git commit -m "WIP"
   ```

2. 非対話的 rebase で fixup:
   ```bash
   GIT_SEQUENCE_EDITOR="sed -i '' '2,\$s/^pick/fixup/'" git rebase -i "$SQUASH_BASE"
   ```
   - コンフリクト時: `git rebase --abort` で中止し、ユーザーに手動解決を促す

3. 差分を分析してコミットメッセージを自動生成:
   ```bash
   git diff $SQUASH_BASE..HEAD
   ```
   Conventional Commits 形式で生成（feat:, fix:, refactor:, docs:, chore:, test: 等）。

4. 生成メッセージでコミットを上書き:
   ```bash
   git commit --amend -m "生成したメッセージ"
   ```

5. 生成したメッセージをユーザーに提示し確認を得る。

#### ケースB: 差分コミットなし（新規コミット）

1. `git add -A`
2. `git diff --cached` で差分確認
3. 差分から Conventional Commits 形式でメッセージ生成
4. `git commit -m "生成したメッセージ"`

### 6. Force Push

```bash
git push --force-with-lease origin <current-branch>
```

失敗時はエラーメッセージを表示。

### 7. 既存PRの確認

`mcp__plugin_github_github__list_pull_requests` を使用:

```
owner: OWNER
repo: REPO
head: "OWNER:<current-branch>"
state: "open"
```

結果が空なら新規作成、PRが見つかれば更新。

### 8. PRテンプレートの検出

```bash
gh repo view --json pullRequestTemplates -q '.pullRequestTemplates'
```

- 1つ → その内容を BODY に使用
- 複数 → AskUserQuestion でユーザーに選択を促す
- なし → デフォルト構造を生成

### 9. PR作成または更新

#### 新規作成

`mcp__plugin_github_github__create_pull_request` を使用:

```
owner: OWNER
repo: REPO
title: コミットメッセージの1行目
head: <current-branch>
base: $PR_BASE
draft: true
body: テンプレート or デフォルト構造
```

**デフォルト body 構造**（テンプレートなしの場合）:

```markdown
## Summary
<差分から読み取った変更内容の要約>

## Changes
- file1.ts
- file2.ts
```

#### 既存PRの更新

`mcp__plugin_github_github__update_pull_request` を使用:

```
owner: OWNER
repo: REPO
pullNumber: 既存PRの番号
title: コミットメッセージの1行目
body: 更新した内容
```

force push 済みなのでコミットは反映済み。

## 注意事項

- `--force-with-lease` で安全な force push
- rebase コンフリクト時は `git rebase --abort` してユーザーに通知
- MCP ツールの owner/repo は `git remote get-url origin` から必ず取得
- PR body にバッククォートや特殊文字を含めても MCP は構造化パラメータなので問題なし
- 親ブランチ検出は3段階フォールバック: tracking config → merge-base距離比較 → デフォルトブランチ
- `$SQUASH_BASE`（コミット集約基準、merge-base ハッシュ）と `$PR_BASE`（PR先ブランチ名）は異なる場合がある
- 親ブランチがリモートに未pushの場合、PR base はデフォルトブランチにフォールバックする（ユーザーに通知）
