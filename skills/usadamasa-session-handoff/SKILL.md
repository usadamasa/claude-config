---
name: session-handoff
description: >
  コンテキスト逼迫時の自動引き継ぎとcompaction。セッション中にコンテキストが逼迫してきたと感じたら、
  プロアクティブにこのスキルを実行する。引継書をauto memoryに保存し、/compactでコンテキストを圧縮する。
  「引き継ぎして」「セッション引き継ぎ」「中断するから引き継ぎ書いて」「handoff」等の手動トリガーにも対応。
---

# session-handoff

コンテキストが逼迫してきたときに引継書を作成し、auto memoryに保存してからコンテキストを圧縮するスキル。
セッション継続と次セッションへの情報伝達の両方を実現する。

## トリガー条件

- **自動(プロアクティブ)**: コンテキストウィンドウが逼迫してきたと判断したとき
- **手動**: 「引き継ぎして」「セッション引き継ぎ」「中断するから引き継ぎ書いて」「handoff」等

## Context

- Current branch: !`git branch --show-current`
- Git status: !`git status --short`
- Recent commits: !`git log --oneline -5`

## ワークフロー

### Step 1: 現在の状態をキャプチャ

以下の情報を収集する:

```bash
git branch --show-current
git status --short
git log --oneline -5
git diff --stat
```

未コミットの変更ファイル一覧、現在のブランチ、直近のコミットを把握する。

### Step 2: タスク進捗を整理

セッション中の作業を振り返り、以下を整理する:

- 完了したステップ(何を実装/修正したか)
- 未完了のステップ(次にやるべきこと)
- 重要な判断とその理由(なぜその方法を選んだか)
- ブロッカー(あれば)

PLAN.md が存在する場合は、その進捗状態も確認する。

### Step 3: SESSION_HANDOFF.md を生成

以下のテンプレートに沿って引継書を作成する:

```markdown
# セッション引き継ぎ

## タスク
<何をしているか、1-2行で簡潔に>

## 完了
- [x] <完了したステップ1>
- [x] <完了したステップ2>

## 次のアクション
1. <次にやること1>
2. <次にやること2>

## Git状態
- ブランチ: <branch>
- 最新コミット: <hash> <message>
- 未コミット変更: <files or "なし">

## 判断・決定
- <重要な判断とその理由>

## ブロッカー
- <あれば。なければ "なし">
```

### Step 4: Auto Memory に保存

プロジェクトの auto memory ディレクトリに `SESSION_HANDOFF.md` を書き込む。

#### worktree 環境の検出と親 memory への保存

まず `.git` がファイルかディレクトリかを確認し、worktree 環境かどうかを判定する:

```bash
# .git がファイルなら worktree 環境
cat .git
# → "gitdir: /path/to/.git/worktrees/feature" なら worktree
# → ディレクトリなら通常のリポジトリ
```

**worktree 環境の場合、2か所に保存する:**

```bash
# .git ファイルから gitdir を取得
GIT_DIR=$(sed 's/^gitdir: //' .git | tr -d '\n')

# ブランチ名を取得 (親 memory のファイル名に使う)
BRANCH_NAME=$(sed 's|ref: refs/heads/||' "$GIT_DIR/HEAD" | tr -d '\n')

# commondir から親 .git ディレクトリを特定
COMMON_REL=$(cat "$GIT_DIR/commondir" | tr -d '\n')
# 相対パスを絶対パスに変換
COMMON_ABS="$(cd "$GIT_DIR" && cd "$COMMON_REL" && pwd)"
PARENT_ROOT="$(dirname "$COMMON_ABS")"

# パスエンコード: / . _ を - に変換
encode_path() { echo "$1" | tr '/._' '-'; }
WORKTREE_ENC=$(encode_path "$(pwd)")
WORKTREE_MEM="$HOME/.claude/projects/$WORKTREE_ENC/memory"
PARENT_ENC=$(encode_path "$PARENT_ROOT")
PARENT_MEM="$HOME/.claude/projects/$PARENT_ENC/memory"
mkdir -p "$WORKTREE_MEM" "$PARENT_MEM"
```

保存先:
1. **主**: worktree の memory path → `$WORKTREE_MEM/SESSION_HANDOFF.md` (現セッションの自動ロード用)
2. **副**: 親リポジトリの memory path → `$PARENT_MEM/SESSION_HANDOFF_${BRANCH_NAME}.md` (親からの可視性用、親の `SESSION_HANDOFF.md` を上書きしない)

**通常リポジトリの場合の保存先特定方法:**

```bash
# CWD のパスから auto memory のプロジェクトディレクトリを導出する
# 例: /Users/masaru_uchida/src/github.com/usadamasa/dotfile/main
#   → ~/.claude/projects/-Users-masaru-uchida-src-github-com-usadamasa-dotfile-main/memory/SESSION_HANDOFF.md
```

auto memory ディレクトリのパスは、CWD の `/` `.` `_` を `-` に置換したものになる。
memory サブディレクトリが存在しない場合は作成する。

既存の `SESSION_HANDOFF.md` がある場合は上書きする(常に最新の状態を反映)。

### Step 5: コンテキストを圧縮して作業を再開

`/compact` を実行してコンテキストを圧縮する。圧縮後はユーザーの確認を待たずに直ちに次のアクションを実行すること。

圧縮指示に「圧縮後の再開命令」を含める:

```
/compact SESSION_HANDOFF.mdの内容を維持しつつコンテキストを圧縮。圧縮完了後はユーザーへの確認なしに直ちに作業を再開すること。次のアクション: <次にやること1>
```

圧縮が完了したら、SESSION_HANDOFF.md の「次のアクション」の先頭から自動的に作業を再開する。ユーザーに「続けますか?」等の確認は不要。

## 注意事項

- 引継書は簡潔に保つ。詳細すぎると次セッションのコンテキストを圧迫する
- MEMORY.md は更新しない(SESSION_HANDOFF.md は次セッション起動時に自動で読み込まれる)
- PLAN.md がある場合は、そのチェック状態を引継書に反映する
