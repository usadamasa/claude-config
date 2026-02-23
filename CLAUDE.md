# CLAUDE.md

> **Note:** `config/claude/` 内の管理対象ファイルは `~/.claude` へファイルレベルで symlink されます｡
> ランタイムファイル(cache, debug, history 等)は `~/.claude` に直接配置され､リポジトリには含まれません｡

## Conversation Guidelines

### 言語設定

- 会話は常に日本語で行う｡
- コミットメッセージ､スキル定義などリポジトリに保存されるテキストは標準語で記載する｡(ConfigのLanguage設定に関わらず)

### 文字種ルール

以下の文字は全角ではなく半角を使用する:

- 句読点: ｡ ､
- 括弧: ( ) [ ] { }
- 記号: , : ; ! ?

## Development Philosophy

### Plan First

Read PLAN.md and implement every step sequentially. After each step, mark it complete with [x] in the plan file. Run all relevant CI checks (ruff, tflint, go vet) after completing all steps. If any check fails, fix it before proceeding. When all steps are done, commit and create a PR.

- プランはあくまで「方針」であり､実装時にコードや設定ファイルを読んで実態との乖離に気づいたら､プランに盲従せず実態を優先すること｡
- 乖離に気づいた時点でユーザーに確認するか､明らかにプラン側の誤りであれば自分で修正して進める｡検証ステップまで問題を先送りしない｡

### Infrastructure changes

Before proposing any infrastructure changes, confirm:

- 1) What is the deployment target (GCE/GKE/Cloud Run)?
- 2) Are there org policies that restrict public IPs, external access, or specific GCP services?
- 3) What auth method is used (service account, IAM, workload identity)?
- 4) List any known constraints from previous failed attempts.

### CI 優先ポリシー

- CI が失敗したら、進行中のタスクに関係ない不具合であっても **即座にすべてのタスクを止めて CI 修正に注力** する
- 「main でも同じだから既存問題」として無視して先に進めてはいけない
- CI が安定して通ることを確認してから元のタスクを再開する

### Test-Driven Development (TDD)

- 原則としてテスト駆動開発(TDD)で進める
- 期待される入出力に基づき､まずテストを作成する
- 実装コードは書かず､テストのみを用意する
- テストを実行し､失敗を確認する
- テストが正しいことを確認できた段階でコミットする
- その後､テストをパスさせる実装を進める
- 実装中はテストを変更せず､コードを修正し続ける
- すべてのテストが通過するまで繰り返す

## Git 操作のプリフライトチェック

git 操作 (commit, push, PR作成など) を行う前に、**必ず以下の環境チェックを実行する**:

1. `cat .git` で worktree 環境かどうかを判定する
   - ファイルで `gitdir: ...` が返る → **worktree 環境**
   - ディレクトリとして存在する → 通常のリポジトリ
2. worktree 環境の場合、`git config --get remote.origin.fetch` を確認する
   - 空なら `git config remote.origin.fetch "+refs/heads/*:refs/remotes/origin/*"` で修正してから `git fetch origin` を実行する
3. `gh pr create` は worktree + bare 環境では `--head {branch_name}` フラグを付ける

## セッション管理

### セッション分割

- 1タスク = 1セッション｡タスク完了後は新規セッションを開始する｡
- plan → implement → verify のサイクルをworktree単位で分離する｡
- 長時間セッションではコンテキストが蓄積し､APIコール1回あたりのinput tokensが増大するため､定期的にセッションを区切る｡

### subagentモデル指定

- Taskツールでの探索(`subagent_type=Explore`)は `model: "haiku"` または `model: "sonnet"` を指定する｡
- 実装の計画(`subagent_type=Plan`)は `model: "opus"` を指定する｡
- 実装は親セッションのデフォルトモデルを使用する｡

## 技術調査とツール

- 技術要素やソフトウェアエンジニアリングについて調査するときは subagent: orm-discovery-mcp-go:oreilly-researcher を積極的に利用するようにしてください｡

## Skills 実装

<https://code.claude.com/docs/en/skills> を参照し､適切な形式で記述してください｡
作成後､Skillsとして利用可能であることを `/skills` で検証してください｡

### スキルの配置場所

| スコープ | 配置場所 | 説明 |
| --------- | --------- | ------ |
| グローバル | `config/claude/skills/` (= `~/.claude/skills/`) | 全プロジェクトで利用可能 |
| プロジェクト | プロジェクトの `.claude/skills/` | そのプロジェクトのみで利用可能 |

特に指示がなければグローバルスコープ(`config/claude/skills/`)に配置してください｡
