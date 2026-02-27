#!/usr/bin/env bash
# baseline.sh - Go プロジェクトのアーキテクチャメトリクス ベースライン測定スクリプト
# 使用方法: baseline.sh <project-root>
# 出力: メトリクスサマリと JSON 詳細ファイル (baseline-YYYYMMDD.json)
set -euo pipefail

PROJECT_ROOT="${1:?使用方法: baseline.sh <project-root>}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_JSON="${PROJECT_ROOT}/baseline-${TIMESTAMP}.json"

if [[ ! -d "$PROJECT_ROOT" ]]; then
    echo "エラー: ディレクトリが存在しません: $PROJECT_ROOT" >&2
    exit 1
fi

if [[ ! -f "${PROJECT_ROOT}/go.mod" ]]; then
    echo "エラー: go.mod が見つかりません。Go プロジェクトルートを指定してください: $PROJECT_ROOT" >&2
    exit 1
fi

echo "=== Go アーキテクチャメトリクス ベースライン測定 ==="
echo "プロジェクト: $PROJECT_ROOT"
echo "測定日時: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# ツールの存在チェック
check_tool() {
    local tool="$1"
    if ! command -v "$tool" &>/dev/null; then
        echo "警告: $tool が見つかりません。スキップします。" >&2
        echo "  インストール方法: aqua.yaml に追加して 'aqua install' を実行してください" >&2
        return 1
    fi
    return 0
}

# aqua が利用可能かチェック
if ! command -v aqua &>/dev/null; then
    echo "情報: aqua が未インストールです。"
    echo "  インストール: brew install aquaproj/aqua/aqua"
    echo "  その後: aqua install"
    echo ""
fi

# golangci-lint の実行
GOLANGCI_RESULT="{}"
if check_tool golangci-lint; then
    echo "--- golangci-lint (テスト可能性・保守性) ---"
    pushd "$PROJECT_ROOT" >/dev/null

    # JSON 形式で実行 (エラーは無視して違反数を収集)
    # v2 では --out-format json は廃止。--output.json.path stdout を使う
    GOLANGCI_JSON=$(golangci-lint run --output.json.path stdout --timeout 5m ./... 2>/dev/null || true)

    if [[ -n "$GOLANGCI_JSON" ]]; then
        # linter 別の違反数を集計
        GOLANGCI_SUMMARY=$(echo "$GOLANGCI_JSON" | \
            python3 -c "
import sys, json
data = json.load(sys.stdin)
issues = data.get('Issues') or []
counts = {}
for issue in issues:
    linter = issue.get('FromLinter', 'unknown')
    counts[linter] = counts.get(linter, 0) + 1
total = sum(counts.values())
print(f'合計違反数: {total}')
for linter, count in sorted(counts.items(), key=lambda x: -x[1]):
    print(f'  {linter}: {count}')
" 2>/dev/null || echo "  集計スキップ (Python3 が必要)")
        echo "$GOLANGCI_SUMMARY"
        GOLANGCI_RESULT="$GOLANGCI_JSON"
    else
        echo "  違反なし (または実行エラー)"
    fi

    popd >/dev/null
    echo ""
fi

# go-arch-lint の実行
ARCH_RESULT="{}"
if check_tool go-arch-lint; then
    echo "--- go-arch-lint (モジュール性・依存方向) ---"
    pushd "$PROJECT_ROOT" >/dev/null

    if [[ ! -f ".go-arch-lint.yml" ]]; then
        echo "  警告: .go-arch-lint.yml が見つかりません。設定ファイルを作成してください。"
        echo "  参考: ~/.claude/skills/usadamasa-go-arch-metrics/references/arch-lint-config.md"
    else
        ARCH_JSON=$(go-arch-lint check --json-output ./... 2>/dev/null || true)
        if [[ -n "$ARCH_JSON" ]]; then
            ARCH_VIOLATIONS=$(echo "$ARCH_JSON" | \
                python3 -c "
import sys, json
data = json.load(sys.stdin)
violations = data.get('violations') or []
print(f'依存方向違反数: {len(violations)}')
for v in violations[:10]:  # 最初の10件のみ表示
    pkg = v.get('packageName', '?')
    dep = v.get('dependencyName', '?')
    print(f'  {pkg} -> {dep}')
if len(violations) > 10:
    print(f'  ... 他 {len(violations) - 10} 件')
" 2>/dev/null || echo "  集計スキップ (Python3 が必要)")
            echo "$ARCH_VIOLATIONS"
            ARCH_RESULT="$ARCH_JSON"
        else
            echo "  依存方向違反なし"
        fi
    fi

    popd >/dev/null
    echo ""
fi

# パッケージ統計 (go list)
echo "--- パッケージ統計 ---"
pushd "$PROJECT_ROOT" >/dev/null
PKG_COUNT=$(go list ./... 2>/dev/null | wc -l | tr -d ' ')
echo "  パッケージ数: $PKG_COUNT"

# テストカバレッジ (参考値)
if command -v go &>/dev/null; then
    echo "  テストカバレッジを測定中..."
    COVERAGE=$(go test ./... -cover 2>/dev/null | \
        grep -oE '[0-9]+\.[0-9]+%' | \
        python3 -c "
import sys
values = [float(x.strip('%')) for x in sys.stdin.read().split() if x.strip('%').replace('.','').isdigit()]
if values:
    print(f'  平均カバレッジ: {sum(values)/len(values):.1f}%')
else:
    print('  カバレッジ: 測定不可')
" 2>/dev/null || echo "  カバレッジ: 測定スキップ")
    echo "$COVERAGE"
fi

popd >/dev/null
echo ""

# JSON サマリファイルの出力
python3 -c "
import json, sys
from datetime import datetime

result = {
    'timestamp': '$(date -u +%Y-%m-%dT%H:%M:%SZ)',
    'project': '$PROJECT_ROOT',
    'golangci_lint': ${GOLANGCI_RESULT},
    'go_arch_lint': ${ARCH_RESULT},
}
with open('$OUTPUT_JSON', 'w') as f:
    json.dump(result, f, indent=2, ensure_ascii=False)
print(f'詳細結果を保存しました: $OUTPUT_JSON')
" 2>/dev/null || echo "JSON 出力スキップ (Python3 が必要)"

echo ""
echo "=== 測定完了 ==="
echo ""
echo "次のステップ:"
echo "  1. 違反を分析: ~/.claude/skills/usadamasa-go-arch-metrics/references/remediation.md"
echo "  2. 設定ファイルを調整: references/golangci-config.md, references/arch-lint-config.md"
echo "  3. CI に統合: references/ci-integration.md"
