---
name: webfetch-domain-manager
description: >-
  settings.jsonのWebFetch/Fetchドメイン許可リストとsandboxネットワーク設定を管理するスキル。
  セッションJSONLから過去30日間のWebFetch/Fetch使用状況を集計し、
  安全性評価に基づいてドメインの追加/削除を提案する。
  (1) 許可ドメインの棚卸し (2) 新規ドメインの追加提案
  (3) 未使用ドメインの削除提案 (4) sandboxドメインの同期管理 に使用する。
---

# WebFetch Domain Manager

settings.jsonの`permissions.allow`に登録されているWebFetch/Fetchドメインの許可リストと、`sandbox.network.allowedDomains`のネットワーク制限を管理するワークフロー。

## ワークフロー

### 1. 分析の実行

以下のコマンドを実行してWebFetch/Fetch使用状況を集計する:

```bash
go run ./cmd/analyze-webfetch --days 30
```

オプション:
- `--days N`: 集計期間を指定(デフォルト: 30日)
- `--settings PATH`: settings.jsonのパスを指定(デフォルト: ~/.claude/settings.json)

### 2. 結果の確認

出力はJSON形式で以下のセクションを含む:

- **metadata**: 分析の概要(期間、ファイル数、WebFetch/Fetch呼び出し回数)
- **current_allowlist**: 現在のpermissions許可ドメイン一覧
- **current_sandbox**: 現在のsandbox.network.allowedDomains一覧
- **recommendations.add**: 追加推奨ドメイン(safeカテゴリで未登録のもの)
- **recommendations.review**: 要確認ドメイン(medium/reviewカテゴリ)
- **recommendations.unused**: 未使用ドメイン(許可リストにあるが使用されていない)
- **recommendations.add_to_sandbox**: sandbox追加推奨(permissionsにあるがsandboxに未登録)
- **all_domains**: 全ドメインの使用統計

### 3. settings.jsonの更新

ユーザーの承認を得た上で、以下の手順でsettings.jsonを更新する:

#### permissions.allowの更新

1. 承認されたドメインのみ`permissions.allow`に追加
2. `WebFetch(domain:ドメイン名)` または `Fetch(domain:ドメイン名)` 形式で記述
3. 許可エントリはアルファベット順にソート
4. 不要と判断されたドメインは削除

#### sandbox.network.allowedDomainsの更新

1. `add_to_sandbox`の推奨に基づき`sandbox.network.allowedDomains`に追加
2. ドメインはアルファベット順にソート
3. ワイルドカード(`*.example.com`)でサブドメインをまとめて許可可能

### 注意事項

- 必ずユーザーの承認を得てから変更を適用すること
- reviewカテゴリのドメインは特に慎重に確認すること
- ワイルドカード(`*.example.com`)は必要な場合のみ使用すること
- permissionsとsandboxは独立した防御層: permissionsはClaude Codeのツール実行権限、sandboxはOSレベルのネットワーク分離

## パーミッション評価順序と ask 配列の注意点

Claude Codeのパーミッション評価は **deny → ask → allow** の順序で行われる｡`ask`は`allow`より優先されるため、`ask`配列の設定によってはドメイン別の`allow`が無効化される｡

### ベア(修飾子なし)エントリの禁止

`ask`配列にベアの`WebFetch`(ドメイン修飾子なし)を入れてはいけない｡

**誤った設定:**
```json
"ask": ["WebFetch"],
"allow": ["WebFetch(domain:github.com)", "WebFetch(domain:docs.anthropic.com)"]
```

この場合、ベアの`WebFetch`がすべてのWebFetch呼び出しにマッチし、`allow`のドメイン別許可が全て無視される｡結果、毎回確認ダイアログが表示される｡

**正しい設定:**
```json
"allow": ["WebFetch(domain:github.com)", "WebFetch(domain:docs.anthropic.com)"]
```

`ask`からベアの`WebFetch`を削除すれば、`allow`に登録されたドメインは確認なしで許可され、未登録ドメインはデフォルト動作(確認を求める)となる｡

### このスキル実行時の検証

ドメイン管理の更新時に、以下を必ず検証すること:

1. `ask`配列にベアの`WebFetch`または`Fetch`が含まれていないか確認する
2. 含まれている場合はユーザーに警告し、削除を提案する
3. `ask`にドメイン指定の`WebFetch(domain:...)`がある場合は意図的な設定なので問題ない

## 安全性カテゴリ

| カテゴリ | 説明 | 例 |
|----------|------|-----|
| safe | 公式ドキュメント・パッケージレジストリ | docs.*, github.com, pkg.go.dev |
| medium | コミュニティ・学習プラットフォーム | stackoverflow.com, medium.com |
| review | 手動確認が必要 | 上記に該当しないドメイン |
