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

# トップレベルディレクトリを列挙 (.で始まるもの除外)
echo "--- 実際のディレクトリ一覧 ---"
declare -A actual_dirs
while IFS= read -r dir; do
    dirname=$(basename "$dir")
    # .で始まるディレクトリを除外
    if [[ "$dirname" == .* ]]; then
        continue
    fi
    file_count=$(find "$dir" -type f -name '*.md' 2>/dev/null | wc -l | tr -d ' ')
    actual_dirs["${dirname}/"]="$file_count"
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
declare -A documented_dirs
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
        dir_name=$(echo "$line" | sed -n 's/.*`\([^`]*\/\)`.*/\1/p')
        if [[ -n "$dir_name" ]]; then
            documented_dirs["$dir_name"]=1
            echo "  $dir_name"
        fi
    fi
done < "$CLAUDE_MD"

echo ""

# 差分を計算
echo "--- 差分レポート ---"
has_diff=false

# 実在するが未記載のディレクトリ
for dir in "${!actual_dirs[@]}"; do
    if [[ -z "${documented_dirs[$dir]:-}" ]]; then
        echo "  [未記載] $dir (${actual_dirs[$dir]} files) - CLAUDE.mdに記載なし"
        has_diff=true
    fi
done

# 記載あるが存在しないディレクトリ
for dir in "${!documented_dirs[@]}"; do
    if [[ -z "${actual_dirs[$dir]:-}" ]]; then
        echo "  [不存在] $dir - CLAUDE.mdに記載あるが実際には存在しない"
        has_diff=true
    fi
done

if [[ "$has_diff" == false ]]; then
    echo "  差分なし - CLAUDE.mdと実態が一致しています"
fi

echo ""
echo "=== 分析完了 ==="
