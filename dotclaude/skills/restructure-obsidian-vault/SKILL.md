---
name: restructure-obsidian-vault
description: Obsidian Vaultのディレクトリ構造を分析し、CLAUDE.mdのディレクトリ構成テーブルを実態に合わせて更新する。「Vaultを整理して」「CLAUDE.mdを更新して」「ディレクトリ構成を確認して」のように構造の棚卸しやドキュメント整備を依頼されたときに使用する。
---

# Obsidian Vault 構造分析・整備

## Overview

Obsidian Vaultの実際のディレクトリ構成とCLAUDE.mdの記載を比較し、差分を検出して更新する。PKMベストプラクティスに基づいた構造改善の提案も行う。

## ワークフロー

### Step 1: 現状分析

`scripts/analyze-vault.sh` を実行してVaultの現状を把握する。

```bash
bash scripts/analyze-vault.sh <vault-path>
```

出力内容:
- トップレベルディレクトリ一覧とファイル数
- CLAUDE.mdに記載されているディレクトリ一覧
- 差分レポート (未記載/不存在)

### Step 2: 差分の確認

分析結果をユーザーに提示し、以下を確認する:
- 未記載ディレクトリの用途 (ユーザーに確認)
- 不存在ディレクトリの扱い (テーブルから削除するか)

### Step 3: CLAUDE.md 更新

ユーザーの確認を得たら、CLAUDE.mdのディレクトリ構成テーブルを更新する。

更新時の注意:
- テーブルの既存フォーマットを維持する (`| ディレクトリ | 用途 |` 形式)
- ディレクトリ名は `` `dirname/` `` 形式でバッククォートで囲む
- 数字プレフィックスのあるディレクトリは数字順にソート
- 数字プレフィックスのないディレクトリはアルファベット順→日本語順

### Step 4: 構造改善の提案 (任意)

ユーザーが希望する場合、`references/pkm-best-practices.md` を参照して以下を提案:
- PARA方式に基づくカテゴリ分類
- フォルダ命名規則の改善
- ディレクトリ移動時の安全性チェック

## リソース

### scripts/

- `analyze-vault.sh`: Vault構造分析スクリプト。トップレベルディレクトリの列挙、ファイル数カウント、CLAUDE.mdとの差分検出を行う。

### references/

- `pkm-best-practices.md`: PKM (Personal Knowledge Management) のベストプラクティス参考資料。PARA方式、命名規則、Obsidian固有の考慮事項をまとめている。
