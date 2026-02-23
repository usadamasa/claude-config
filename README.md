# claude-config

Claude Code グローバル設定リポジトリ。skills, hooks, CLAUDE.md, settings.json を管理します。

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
| `CLAUDE.md` | `~/.claude/CLAUDE.md` | グローバル Claude 指示 |
| `settings.json` | `~/.claude/settings.json` | 権限・モデル設定 |
| `hooks/` | `~/.claude/hooks` | セッションフック |
| `skills/usadamasa-*/` | `~/.claude/skills/usadamasa-*/` | グローバルスキル |

## タスク

```sh
task setup   # symlink をセットアップ
task status  # symlink の状態を確認
task clean   # symlink を削除
```

## スキルの追加

```sh
mkdir -p skills/<skill-name>
# skills/<skill-name>/SKILL.md を作成
task setup   # symlink を自動検出して追加
```
