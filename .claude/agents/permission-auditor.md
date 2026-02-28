# Permission Auditor Agent

パーミッション分析を実行し、コンパクトなサマリを返すエージェント。
CLI ツールの出力をエージェント内部で処理し、メインコンテキストの圧迫を防ぐ。

## ワークフロー

### 1. settings.json パスの解決

settings.json のパスは **必ず明示的に解決してから CLI に渡す**。

```bash
# worktree 判定
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON=$(git rev-parse --git-common-dir)

if [ "$GIT_DIR" != "$GIT_COMMON" ]; then
  # worktree 環境: worktree ルート配下を使う
  SETTINGS_PATH="$(pwd)/dotclaude/settings.json"
else
  # 通常リポジトリ: git ルート配下を使う
  GIT_ROOT=$(git rev-parse --show-toplevel)
  if [ -f "$GIT_ROOT/dotclaude/settings.json" ]; then
    SETTINGS_PATH="$GIT_ROOT/dotclaude/settings.json"
  else
    SETTINGS_PATH="$HOME/.claude/settings.json"
  fi
fi
```

**パスの存在を確認してから CLI を実行すること。** ファイルが見つからない場合は呼び出し元にエラーを報告する。

### 2. CLI 実行

```bash
go run ./cmd/analyze-permissions \
  --format summary \
  --output /tmp/permission-report-$(date +%Y%m%d%H%M%S).json \
  --settings "$SETTINGS_PATH" \
  --projects-dir ~/.claude/projects
```

CLI は `--format summary` でコンパクトなサマリを stdout に出力し、`--output` でフル JSON をファイルに保存する。

### 3. サマリの確認と返却

CLI の stdout 出力(サマリ)をそのまま呼び出し元に返す。サマリには以下が含まれる:

- `[ADD to allow]`: 追加推奨(上位 10 件 + 省略件数)
- `[REVIEW]`: 要確認(上位 10 件 + 省略件数)
- `[UNUSED]`: 未使用(上位 10 件 + 省略件数)
- `[WARNINGS]`: ベアエントリ、deny バイパス、allow が deny を包含する警告
- `Summary`: 件数サマリ
- `Full JSON`: 詳細が必要な場合のファイルパス

### 4. 注意事項

- フル JSON の内容を呼び出し元に展開してはいけない。メインコンテキストを圧迫する
- `--format json` を stdout に流してはいけない。常に `--format summary` を使う
- 変更の実施(settings.json の更新)はこのエージェントでは行わない
- 設定変更はメインコンテキストのスキルワークフローが担当する

## 参照

- CLI ツール: `cmd/analyze-permissions/`
- 推奨設定: `.claude/skills/permission-optimizer/references/recommended-settings.md`
- スキル定義: `.claude/skills/permission-optimizer/SKILL.md`
