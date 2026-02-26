# Claude Code 設定リポジトリ

Claude Code のグローバル設定を構成管理するリポジトリ｡
`CLAUDE-global.md`, `settings.json`, `skills/`, `hooks/` 等を `~/.claude/` へ symlink して適用する｡

## ディレクトリ構成

| パス | 説明 |
| ---- | ---- |
| `CLAUDE-global.md` | グローバル CLAUDE.md (→ `~/.claude/CLAUDE.md`) |
| `CLAUDE.md` | このファイル (プロジェクトスコープ) |
| `settings.json` | パーミッション・モデル設定 (→ `~/.claude/settings.json`) |
| `env.sh` | 環境変数設定 (→ `~/.claude/env.sh`) |
| `hooks/` | セッションフック (→ `~/.claude/hooks/`) |
| `skills/` | グローバルスキル (各サブディレクトリ → `~/.claude/skills/`) |
| `cmd/` | Go CLI ツール (詳細は `cmd/CLAUDE.md`) |
| `tests/` | bats テスト |
| `Taskfile.yml` | タスクランナー定義 |

## タスク実行

```sh
task setup   # symlink をセットアップ
task status  # symlink の状態を確認
task clean   # symlink を削除
task test    # bats + Go テストを統合実行
task go:test # Go テストのみ実行
```

## スキル開発

1. `skills/<skill-name>/` ディレクトリを作成
2. `skills/<skill-name>/SKILL.md` にスキル定義を記述
3. `task setup` で `~/.claude/skills/<skill-name>` に自動リンク
4. `/skills` で利用可能か検証

## symlink の仕組み

`task setup` は以下の symlink を作成する:

- `CLAUDE-global.md` → `~/.claude/CLAUDE.md` (名前が異なる特殊マッピング)
- `settings.json` → `~/.claude/settings.json`
- `env.sh` → `~/.claude/env.sh`
- `hooks/` → `~/.claude/hooks/`
- `skills/*/` → `~/.claude/skills/*/` (各スキルディレクトリを個別リンク)

## worktree 環境での注意事項

このリポジトリは bare リポジトリ + git worktree で運用される場合がある｡
worktree 環境では `.git` がファイル(ディレクトリではない)になるため､
`cat .git` で `gitdir:` が返るかどうかで判定する｡
