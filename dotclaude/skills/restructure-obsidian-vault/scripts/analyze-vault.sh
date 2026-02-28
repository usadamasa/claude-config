#!/usr/bin/env bash
# analyze-vault.sh - Obsidian Vaultのディレクトリ構造を分析し、CLAUDE.mdとの差分を表示する
set -euo pipefail

VAULT_PATH="${1:?Usage: analyze-vault.sh <vault-path>}"
CLAUDE_MD="${VAULT_PATH}/CLAUDE.md"

if [[ ! -d "$VAULT_PATH" ]]; then
    echo "Error: Vault path does not exist: $VAULT_PATH" >&2
    exit 1
fi

echo "=== Obsidian Vault 構造分析 ==="
echo "Vault: $VAULT_PATH"
echo ""

# 一時ファイル
actual_list=$(mktemp)
documented_list=$(mktemp)
actual_detail=$(mktemp)
trap 'rm -f "$actual_list" "$documented_list" "$actual_detail"' EXIT

# トップレベルディレクトリを列挙 (.で始まるもの除外)
echo "--- 実際のディレクトリ一覧 ---"
while IFS= read -r dir; do
    dirname=$(basename "$dir")
    # .で始まるディレクトリを除外
    if [[ "$dirname" == .* ]]; then
        continue
    fi
    file_count=$(find "$dir" -type f -name '*.md' 2>/dev/null | wc -l | tr -d ' ')
    echo "${dirname}/" >> "$actual_list"
    printf "%s/\t%s\n" "$dirname" "$file_count" >> "$actual_detail"
    printf "  %-30s %4d files\n" "${dirname}/" "$file_count"
done < <(find "$VAULT_PATH" -maxdepth 1 -mindepth 1 -type d | sort)

echo ""

# CLAUDE.mdからディレクトリ名を抽出
if [[ ! -f "$CLAUDE_MD" ]]; then
    echo "Warning: CLAUDE.md not found at $CLAUDE_MD" >&2
    echo "差分チェックをスキップします"
    exit 0
fi

echo "--- CLAUDE.md 記載ディレクトリ ---"
in_table=false
while IFS= read -r line; do
    # ディレクトリ構成テーブルの開始を検出
    if echo "$line" | grep -q '| ディレクトリ | 用途 |'; then
        in_table=true
        continue
    fi
    # テーブル区切り行をスキップ
    if echo "$line" | grep -q '^|---|'; then
        continue
    fi
    # テーブル外に出たら終了 (空行やテーブル以外の行)
    if $in_table && ! echo "$line" | grep -q '^|'; then
        in_table=false
        continue
    fi
    # テーブル行からディレクトリ名を抽出 (| `dirname/` | ... | 形式)
    if $in_table; then
        # shellcheck disable=SC2016 # Backticks are literal in the sed pattern
        dir_name=$(echo "$line" | sed -n 's/.*`\([^`]*\/\)`.*/\1/p')
        if [[ -n "$dir_name" ]]; then
            echo "$dir_name" >> "$documented_list"
            echo "  $dir_name"
        fi
    fi
done < "$CLAUDE_MD"

echo ""

# ソートして差分を計算
sort -o "$actual_list" "$actual_list"
sort -o "$documented_list" "$documented_list"

echo "--- 差分レポート ---"
has_diff=false

# 実在するが未記載のディレクトリ (actual_list にのみ存在)
while IFS= read -r dir; do
    file_count=$(awk -F'\t' -v d="$dir" '$1 == d {print $2}' "$actual_detail")
    echo "  [未記載] $dir (${file_count} files) - CLAUDE.mdに記載なし"
    has_diff=true
done < <(comm -23 "$actual_list" "$documented_list")

# 記載あるが存在しないディレクトリ (documented_list にのみ存在)
while IFS= read -r dir; do
    echo "  [不存在] $dir - CLAUDE.mdに記載あるが実際には存在しない"
    has_diff=true
done < <(comm -13 "$actual_list" "$documented_list")

if [[ "$has_diff" == false ]]; then
    echo "  差分なし - CLAUDE.mdと実態が一致しています"
fi

echo ""
echo "=== 分析完了 ==="
