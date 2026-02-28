#!/usr/bin/env bash
# obs-vault-health.sh - Obsidian Vault 健全性チェックレポート
# Usage: obs-vault-health.sh [vault=<name>]
set -euo pipefail

# vault= オプションを透過的に渡す (配列で安全に引数を扱う)
VAULT_ARGS=()
for arg in "$@"; do
  case "$arg" in
    vault=*) VAULT_ARGS+=("$arg") ;;
  esac
done

# single-value コマンド (total, info=) の値を取得するヘルパー
obs_val() {
  obsidian "$@" 2>/dev/null
}

vault_name=$(obs_val vault "${VAULT_ARGS[@]}" info=name || echo "unknown")
file_count=$(obs_val files "${VAULT_ARGS[@]}" total || echo "?")
folder_count=$(obs_val folders "${VAULT_ARGS[@]}" total || echo "?")
orphan_count=$(obs_val orphans "${VAULT_ARGS[@]}" total || echo "?")
deadend_count=$(obs_val deadends "${VAULT_ARGS[@]}" total || echo "?")
unresolved_count=$(obs_val unresolved "${VAULT_ARGS[@]}" total || echo "?")
task_todo=$(obs_val tasks todo "${VAULT_ARGS[@]}" total || echo "?")
task_done=$(obs_val tasks done "${VAULT_ARGS[@]}" total || echo "?")
tag_count=$(obs_val tags "${VAULT_ARGS[@]}" total || echo "?")

# orphan/deadend 率を計算
if [[ "$file_count" =~ ^[0-9]+$ ]] && [ "$file_count" -gt 0 ]; then
  if [[ "$orphan_count" =~ ^[0-9]+$ ]]; then
    orphan_pct=$(( orphan_count * 100 / file_count ))
  else
    orphan_pct="?"
  fi
  if [[ "$deadend_count" =~ ^[0-9]+$ ]]; then
    deadend_pct=$(( deadend_count * 100 / file_count ))
  else
    deadend_pct="?"
  fi
else
  orphan_pct="?"
  deadend_pct="?"
fi

# 判定
judge() {
  local val=$1 warn=$2 crit=$3
  if ! [[ "$val" =~ ^[0-9]+$ ]]; then
    echo "?"
    return
  fi
  if [ "$val" -le "$warn" ]; then
    echo "OK"
  elif [ "$val" -le "$crit" ]; then
    echo "WARN"
  else
    echo "CRIT"
  fi
}

orphan_status=$(judge "${orphan_pct}" 10 20)
deadend_status=$(judge "${deadend_pct}" 15 30)
unresolved_status=$(judge "${unresolved_count}" 0 10)

echo "=== Vault Health Report: ${vault_name} ==="
echo ""
echo "--- 概要 ---"
echo "ファイル数:     ${file_count}"
echo "フォルダ数:     ${folder_count}"
echo "タグ数:         ${tag_count}"
echo ""
echo "--- タスク ---"
echo "未完了:         ${task_todo}"
echo "完了済:         ${task_done}"
echo ""
echo "--- リンク健全性 ---"
echo "孤立ノート:     ${orphan_count} (${orphan_pct}%) [${orphan_status}]"
echo "デッドエンド:   ${deadend_count} (${deadend_pct}%) [${deadend_status}]"
echo "未解決リンク:   ${unresolved_count} [${unresolved_status}]"
echo ""
echo "--- 判定基準 ---"
echo "orphan:     OK=<10% / WARN=10-20% / CRIT=>20%"
echo "deadend:    OK=<15% / WARN=15-30% / CRIT=>30%"
echo "unresolved: OK=0    / WARN=1-10   / CRIT=>10"
