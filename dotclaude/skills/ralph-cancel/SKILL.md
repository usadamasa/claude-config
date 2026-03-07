---
name: ralph-cancel
description: >
  アクティブな Ralph Loop をキャンセルする。
  状態ファイルを削除してループを停止する。
---

# Ralph Cancel コマンド

## 手順

1. Bash で `tmp/ralph-loop.local.md` の存在を確認する:
   ```
   test -f tmp/ralph-loop.local.md && echo "EXISTS" || echo "NOT_FOUND"
   ```

2. **NOT_FOUND の場合**: 「アクティブな Ralph loop はありません。」と報告する

3. **EXISTS の場合**:
   - Read ツールで `tmp/ralph-loop.local.md` を読み、`iteration:` フィールドの値を取得する
   - Bash で状態ファイルを削除する: `rm tmp/ralph-loop.local.md`
   - 「Ralph loop をキャンセルしました (イテレーション N で停止)」と報告する (N は iteration の値)
