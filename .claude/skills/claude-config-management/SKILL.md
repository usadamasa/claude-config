---
name: claude-config-management
description: >-
  Claude Code設定(`config/claude/`)の構成管理ガイド。ファイルレベルsymlinkによる設定管理、
  管理対象の追加・削除、Taskfileタスクの実行方法を提供する。
  「設定ファイルを追加して」「新しいスキルを追加して」「symlinkの状態を確認して」
  「config/claudeを変更して」「.gitignoreを更新して」のように
  Claude Code設定の構成変更を行うときに使用する。
---

# Claude Code 構成管理

## アーキテクチャ

`~/.claude` は実ディレクトリ。`config/claude/` 内の管理対象ファイルだけを個別に symlink する。
ランタイムファイル(cache, debug, history 等)はリポジトリに含まれない。

```
~/.claude/                    (実ディレクトリ)
├── CLAUDE.md              -> config/claude/CLAUDE.md
├── settings.json          -> config/claude/settings.json
├── hooks/                 -> config/claude/hooks/
├── skills/
│   ├── usadamasa-*/       -> config/claude/skills/usadamasa-*/  (symlink)
│   └── <plugin-skills>/      (実ディレクトリ、管理外)
├── cache/                    (ランタイム、管理外)
├── projects/                 (ランタイム、管理外)
└── ...
```

**管理方針**: `config/claude/.gitignore` で除外されていないもの = 管理対象。

## タスク

```sh
task claude:setup    # symlink セットアップ (マイグレーション含む)
task claude:status   # 状態確認
task claude:clean    # symlink 削除
```

タスク定義: `config/claude/Taskfile.yml`
トップレベルからは `claude:` namespace で呼び出される。

## 管理対象の追加手順

### トップレベルファイルを追加

1. `config/claude/<filename>` にファイルを配置
2. `config/claude/Taskfile.yml` の `setup` タスクの `for file in CLAUDE.md settings.json` に追加
3. `config/claude/.gitignore` で除外されていないことを確認
4. `task claude:setup` を実行

### トップレベルディレクトリを追加

1. `config/claude/<dirname>/` にディレクトリを配置
2. `config/claude/Taskfile.yml` の `setup` タスクの `for dir in hooks` に追加
3. `task claude:setup` を実行

### グローバルスキルを追加 (git管理対象)

1. `config/claude/skills/usadamasa-<skill-name>/SKILL.md` を作成
2. `task claude:setup` を実行 (`usadamasa-*` を自動検出、Taskfile変更不要)

### プロジェクトスコープのスキルを追加

`.claude/skills/<skill-name>/SKILL.md` を作成。symlink 不要。

## 注意事項

- `ln -sfn` でディレクトリ先に既存ディレクトリがあるとネスト symlink が発生する。
  `~/.claude` が実ディレクトリであることを前提としている
- `.gitignore` の `!/skills/usadamasa-*/` パターンで管理対象スキルを制御
- `task claude:setup` はべき等 (何度実行しても安全)
