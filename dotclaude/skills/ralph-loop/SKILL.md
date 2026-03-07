---
name: ralph-loop
description: >
  現在のセッションで Ralph Loop を開始する。
  反復的な開発ループにより、同じプロンプトを繰り返しフィードして自己改善を行う。
  複数行プロンプトに対応。
argument-hint: "PROMPT [--max-iterations N] [--completion-promise TEXT]"
---

# Ralph Loop コマンド

## 引数のパース

`$ARGUMENTS` を以下のルールで解析する (シェルに渡さないこと):

1. `--max-iterations N` があれば N を抽出する (デフォルト: 0 = 無制限)
2. `--completion-promise TEXT` があれば TEXT を抽出する (デフォルト: null)
3. 残りすべてがプロンプト本文

プロンプトが空の場合はエラーメッセージを出して終了する。

## 状態ファイルの作成

1. Bash で `mkdir -p tmp` を実行する
2. Write ツールで `tmp/ralph-loop.local.md` を以下の形式で作成する:

```
---
active: true
iteration: 1
session_id: (Bash で `echo $CLAUDE_CODE_SESSION_ID` を実行して取得)
max_iterations: <パースした値>
completion_promise: "<パースした値>" (null の場合は null)
started_at: "<現在のUTC時刻 ISO8601>"
---

<プロンプト本文 (複数行OK)>
```

**重要**: Write ツールを使うことで、複数行プロンプトがそのまま安全に書き込まれる。

## セットアップメッセージ

状態ファイル作成後、以下を出力する:

```
Ralph loop activated in this session!

Iteration: 1
Max iterations: <N or unlimited>
Completion promise: <TEXT or none>
```

## タスク実行

セットアップ後、プロンプトの内容に取り掛かる。

## 重要なルール

- **completion promise が設定されている場合**: `<promise>TEXT</promise>` タグは、その内容が完全かつ明白に TRUE であるときのみ出力すること
- ループを抜けるために嘘の promise を出力してはならない
- 行き詰まったと感じても、promise が真になるまでループを続けること

## 関連コマンド

- `/ralph-cancel` - アクティブなループをキャンセルする
- `/ralph-loop-help` - Ralph Loop の詳しい説明を表示する
