---
name: bats-testing
description: bats-core を使用したシェルスクリプトの単体テスト作成ガイド。テストの書き方、フィクスチャ、モック、TDD の進め方を提供します。dotfiles リポジトリでの実例を含みます。
---

# bats-core シェルスクリプトテストガイド

このスキルは、bats-core を使用したシェルスクリプトの単体テスト作成方法を提供します。

## クイックリファレンス

### インストールと実行

```bash
# インストール
brew install bats-core

# テスト実行
bats tests/                    # 全テスト
bats tests/my-test.bats        # 特定ファイル
bats --verbose-run tests/      # 詳細出力
```

### 基本構文

```bash
#!/usr/bin/env bats

setup() {
  # 各テスト前に実行
  load 'helpers/mock.sh'
  source "$DOTFILE_DIR/script.sh"
}

teardown() {
  # 各テスト後に実行
  cleanup
}

@test "テストの説明" {
  run my_command
  [ "$status" -eq 0 ]
  [[ "$output" == *"expected"* ]]
}
```

### run コマンド

```bash
run my_command arg1 arg2

$status  # 終了ステータス
$output  # 標準出力 + 標準エラー
$lines   # 出力を行ごとに分割した配列
```

**注意**: `run` はサブシェル実行のため、グローバル変数の変更は親に反映されない

### アサーション

```bash
[ "$status" -eq 0 ]                    # 成功
[ "$status" -ne 0 ]                    # 失敗
[[ "$output" == *"substring"* ]]       # 部分一致
[[ "$output" != *"unwanted"* ]]        # 不一致
[ "${#lines[@]}" -eq 5 ]               # 行数
```

## dotfiles での使用箇所

| ファイル | 説明 |
|---------|------|
| `tests/peco-gcop.bats` | peco-gcop 関数のテスト (10 ケース) |
| `tests/fixtures/git-setup.sh` | テスト用 Git リポジトリ作成 |
| `tests/helpers/peco-mock.sh` | peco コマンドのモック |
| `tests/helpers/zle-mock.sh` | zsh zle/bindkey のモック |

### ディレクトリ構造

```
tests/
├── *.bats              # メインテストファイル
├── fixtures/           # テスト用データ・セットアップ
└── helpers/            # モック・ヘルパー関数
```

## TDD ワークフロー

```
1. Red: テストを書く → 失敗を確認 → コミット
2. Green: 実装する → テストパス → コミット
3. Refactor: コード改善 → テスト維持
```

## よくある問題

| 問題 | 解決 |
|-----|------|
| `bindkey: command not found` | モックをテスト対象より先に読み込む |
| 変数が変更されない | `run` の代わりに直接実行 |
| `trap: undefined signal: RETURN` | `trap ... EXIT` を使う (bash 互換) |

## 詳細情報

モックの作成、フィクスチャの詳細、テスタブルなコード設計については [reference.md](reference.md) を参照してください。
