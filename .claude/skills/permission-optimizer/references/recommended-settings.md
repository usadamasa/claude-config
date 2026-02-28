# 推奨パーミッション設定リファレンス

## 設定ファイルの優先順位

Claude Code はパーミッションを以下の優先順位で評価する(高い方が優先):

1. **Managed settings** (最高) - `/Library/Application Support/ClaudeCode/managed-settings.json`
2. **CLI 引数**
3. **Project local** - `.claude/settings.local.json`
4. **Project shared** - `.claude/settings.json`
5. **User global** (最低) - `~/.claude/settings.json`

## パーミッション評価順序

同一設定ファイル内の評価順序: **deny → ask → allow**

- deny にマッチしたら即拒否(allow は無視)
- ask にマッチしたらユーザーに確認
- allow にマッチしたら自動許可
- どれにもマッチしなければツールのデフォルト動作

## ワイルドカード記法

### Bash パーミッション

| パターン | 説明 | 例 |
|----------|------|-----|
| `Bash(cmd:*)` | プレフィックスマッチ | `Bash(git status:*)` |
| `Bash(cmd)` | 完全一致 | `Bash(ls)` |

### Read/Write/Edit パーミッション

| パターン | 説明 | 例 |
|----------|------|-----|
| `Read(path/**)` | ディレクトリ以下全て | `Read(~/.claude/**)` |
| `Read(path)` | ファイル完全一致 | `Read(CLAUDE.md)` |
| `Read(**)` | additionalDirectories 全て | `Read(**)` |

## スコープ別推奨配置

### global (`~/.claude/settings.json`)

プロジェクトを横断して使う汎用コマンド:

```json
{
  "permissions": {
    "allow": [
      "Bash(git status:*)", "Bash(git log:*)", "Bash(git diff:*)",
      "Bash(git branch:*)", "Bash(git fetch:*)", "Bash(git rev-parse:*)",
      "Bash(go:*)", "Bash(task:*)", "Bash(make:*)",
      "Bash(gh pr:*)", "Bash(gh run:*)", "Bash(gh repo:*)", "Bash(gh api:*)",
      "Bash(docker:*)", "Bash(brew:*)",
      "Bash(ls:*)", "Bash(pwd:*)", "Bash(which:*)", "Bash(wc:*)",
      "Read(~/.claude/**)", "Edit(~/.claude/**)", "Write(~/.claude/**)"
    ],
    "deny": [
      "Bash(curl:*)", "Bash(wget:*)",
      "Bash(sudo:*)", "Bash(ssh:*)", "Bash(scp:*)", "Bash(eval:*)",
      "Bash(gh auth:*)",
      "Read(~/.ssh/**)", "Read(~/.aws/**)", "Read(~/.gnupg/**)",
      "Read(~/.kube/**)", "Read(~/.docker/config.json)",
      "Read(~/.zsh_history)", "Read(~/.bash_history)",
      "Write(~/.ssh/**)", "Write(~/.aws/**)", "Write(~/.gnupg/**)"
    ],
    "ask": [
      "Bash(git commit:*)", "Bash(git push:*)",
      "Bash(git rebase:*)", "Bash(git reset:*)",
      "Bash(rm -rf:*)"
    ]
  }
}
```

### project shared (`.claude/settings.json`)

プロジェクト固有のツールやパス:

```json
{
  "permissions": {
    "allow": [
      "Bash(npm:*)", "Bash(npx:*)",
      "Read(src/**)", "Write(src/**)", "Edit(src/**)"
    ]
  }
}
```

### project local (`.claude/settings.local.json`)

一時的・個人的なパーミッション(Git 管理外):

```json
{
  "permissions": {
    "allow": [
      "Bash(gcloud:*)"
    ]
  }
}
```

## curl/wget の取り扱い

**推奨: deny に設定し、WebFetch ツールを使用する**

理由:
- WebFetch はサンドボックス内で動作し、セキュリティが担保される
- `allowedDomains` でアクセス先を制御可能
- curl/wget は任意のデータ送信が可能で、情報漏洩リスクがある

API テスト等で必要な場合は project local で allow に追加する。

## deny バイパスリスク一覧

| Bash コマンド | バイパス対象 | 代替ツール |
|--------------|-------------|-----------|
| `cat`, `head`, `tail` | Read deny | Read ツール |
| `grep`, `awk` | Read deny | Grep ツール |
| `echo`, `tee` | Write deny | Write ツール |
| `cp`, `mv` | Write deny | Write ツール |
| `find` | Read + Write deny | Glob ツール |
| `sed` | Read + Write deny | Edit ツール |

**原則**: Claude Code には専用ツール(Read/Grep/Write/Edit/Glob)があるため、Bash でのファイル操作コマンドは allow に含めない。

## ベアエントリの禁止

`"Bash"` や `"Read"` のような修飾子なしエントリは、全ての呼び出しにマッチするため禁止。

**誤った設定:**
```json
"ask": ["Bash"],
"allow": ["Bash(git status:*)"]
```
→ ベアの `Bash` が全 Bash コマンドにマッチし、allow の個別設定が無意味になる。

## パーミッション集約の注意

スコープ拡大を防ぐ:

```json
// NG: mkdir がファイルシステム全体に適用される
"Bash(mkdir:*)"

// OK: スコープを限定
"Bash(mkdir -p ~/.claude:*)", "Bash(mkdir .claude:*)"
```
