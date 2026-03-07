# mmdc オプションリファレンス

## コマンド構文

```
mmdc [options]
```

## 主要オプション

| オプション | 短縮 | 説明 | デフォルト |
|-----------|------|------|-----------|
| `--input` | `-i` | 入力ファイル (`.mmd`) | stdin |
| `--output` | `-o` | 出力ファイル | stdout |
| `--theme` | `-t` | テーマ (default/dark/forest/neutral) | `default` |
| `--width` | `-w` | 幅 (px) | `800` |
| `--height` | `-H` | 高さ (px) | `600` |
| `--scale` | `-s` | スケール倍率 | `1` |
| `--backgroundColor` | `-b` | 背景色 (CSS色名/hex/transparent) | `white` |
| `--configFile` | `-c` | Mermaid 設定 JSON | - |
| `--cssFile` | `-C` | カスタム CSS | - |
| `--puppeteerConfigFile` | `-p` | Puppeteer 設定 JSON | - |
| `--quiet` | `-q` | ログ抑制 | `false` |

## 設定ファイル

スキル同梱の `config.json` を参照。`themeVariables.fontFamily` で日本語フォントを指定している。

## 推奨プリセット

### Confluence 添付用 (高解像度 PNG)

```bash
mmdc -i input.mmd -o output.png -c ~/.claude/skills/mermaid-cli/config.json -s 2 -w 1200 -b white -t neutral
```

### プレゼンテーション用 (大きめ PNG)

```bash
mmdc -i input.mmd -o output.png -c ~/.claude/skills/mermaid-cli/config.json -s 3 -w 1920 -b transparent
```

### ドキュメント埋め込み用 (SVG)

```bash
mmdc -i input.mmd -o output.svg -c ~/.claude/skills/mermaid-cli/config.json -t default
```

## 対応ダイアグラム

flowchart, sequence, class, state, er, gantt, pie, mindmap, timeline, quadrant, sankey, xy-chart, block
