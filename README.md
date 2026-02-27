# claude-config

Claude Code グローバル設定リポジトリ。`dotclaude/` 配下の skills, hooks, CLAUDE.md, settings.json を管理します。

## セットアップ

```sh
# リポジトリを clone (ghq 推奨)
ghq get github.com/usadamasa/claude-config

# symlink をセットアップ
cd ~/src/github.com/usadamasa/claude-config
task setup
```

## 管理対象ファイル

| ファイル/ディレクトリ | symlink 先 | 説明 |
|---|---|---|
| `dotclaude/CLAUDE-global.md` | `~/.claude/CLAUDE.md` | グローバル Claude 指示 |
| `dotclaude/settings.json` | `~/.claude/settings.json` | 権限・モデル設定 |
| `dotclaude/hooks/` | `~/.claude/hooks` | セッションフック |
| `dotclaude/skills/usadamasa-*/` | `~/.claude/skills/usadamasa-*/` | グローバルスキル |

## タスク

```sh
task setup   # symlink をセットアップ
task status  # symlink の状態を確認
task clean   # symlink を削除
```

## スキルの追加

```sh
mkdir -p dotclaude/skills/<skill-name>
# dotclaude/skills/<skill-name>/SKILL.md を作成
task setup   # symlink を自動検出して追加
```
