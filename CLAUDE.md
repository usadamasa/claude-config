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

## worktree 環境でのファイルパス解決

このリポジトリは通常リポジトリ + git worktree (linked worktree) で運用される場合がある｡
worktree 環境では symlink 対象ファイルの編集先・参照先が通常と異なるため､
以下のプロトコルに従うこと｡

### worktree 判定

`git rev-parse` で判定する:

```bash
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON=$(git rev-parse --git-common-dir)
# GIT_DIR != GIT_COMMON → worktree 環境
# GIT_DIR == GIT_COMMON → 通常リポジトリ
```

### symlink 対象ファイルのパス解決

| ファイル | 通常リポジトリ | worktree 環境 |
| ---- | ---- | ---- |
| `settings.json` | `~/.claude/settings.json` | `$(pwd)/settings.json` |
| `env.sh` | `~/.claude/env.sh` | `$(pwd)/env.sh` |
| `env.sh.example` | `$(pwd)/env.sh.example` | `$(pwd)/env.sh.example` |
| `CLAUDE-global.md` | `~/.claude/CLAUDE.md` | `$(pwd)/CLAUDE-global.md` |

### スキル実行時の手順

1. `git rev-parse --git-dir` と `--git-common-dir` を比較して worktree 判定
2. worktree なら Edit ツールのパスに `$(pwd)/<file>` を使用する
3. CLI ツール (`analyze-permissions`, `analyze-webfetch`) は worktree なら `--settings $(pwd)/settings.json` を付ける

### 検証フロー

worktree で変更した設定を動作確認するフロー:

```
1. worktree で settings.json を編集
2. worktree から `task setup` を実行 → symlink が worktree のファイルを指す
3. 新しい Claude Code セッションで動作確認
4. 確認後、メインリポジトリから `task setup` を実行して復元:
   MAIN_WORKTREE=$(git rev-parse --git-common-dir | xargs dirname)
   (cd "$MAIN_WORKTREE" && task setup)
```

注意: symlink 復元を忘れると､worktree 削除後にリンク切れになる｡
検証完了後は必ずメインリポジトリから `task setup` で復元すること｡
