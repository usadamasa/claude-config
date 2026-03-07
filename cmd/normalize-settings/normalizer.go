package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Normalize は settings.json のバイト列を正規化する｡
// 戻り値: 正規化済みバイト列、警告メッセージ、エラー
func Normalize(data []byte, pinnedModel string, stripFields []string) ([]byte, []string, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, nil, fmt.Errorf("JSON パースに失敗: %w", err)
	}

	var warns []string

	// ランタイムフィールドを除去
	for _, field := range stripFields {
		delete(top, field)
	}

	// model フィールドの検証
	if pinnedModel != "" {
		if raw, ok := top["model"]; ok {
			var model string
			if err := json.Unmarshal(raw, &model); err == nil && model != pinnedModel {
				warns = append(warns, "model が期待値と不一致")
			}
		}
	}

	// permissions 配列のソート
	if raw, ok := top["permissions"]; ok {
		sorted, err := sortPermissions(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("permissions の正規化に失敗: %w", err)
		}
		top["permissions"] = sorted
	}

	// sandbox の正規化
	if raw, ok := top["sandbox"]; ok {
		sorted, err := sortSandbox(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("sandbox の正規化に失敗: %w", err)
		}
		top["sandbox"] = sorted
	}

	// enabledPlugins のキーソート
	if raw, ok := top["enabledPlugins"]; ok {
		sorted, err := sortObjectKeys(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("enabledPlugins の正規化に失敗: %w", err)
		}
		top["enabledPlugins"] = sorted
	}

	out, err := marshalIndent(top)
	if err != nil {
		return nil, nil, fmt.Errorf("JSON シリアライズに失敗: %w", err)
	}

	return out, warns, nil
}

// NormalizeFile はファイルを読み込み、正規化して書き戻す｡
// changed: 内容が変わったか、warns: 警告、err: エラー
func NormalizeFile(path, pinnedModel string, stripFields []string) (bool, []string, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- CLIツール: パスはフラグ引数由来
	if err != nil {
		return false, nil, fmt.Errorf("ファイル読み込みに失敗: %w", err)
	}

	normalized, warns, err := Normalize(data, pinnedModel, stripFields)
	if err != nil {
		return false, nil, err
	}

	if bytes.Equal(data, normalized) {
		return false, warns, nil
	}

	if err := os.WriteFile(path, normalized, 0600); err != nil { // #nosec G306
		return false, nil, fmt.Errorf("ファイル書き込みに失敗: %w", err)
	}

	return true, warns, nil
}

// sortPermissions は permissions オブジェクト内の配列をソートする｡
func sortPermissions(raw json.RawMessage) (json.RawMessage, error) {
	var perms map[string]json.RawMessage
	if err := json.Unmarshal(raw, &perms); err != nil {
		return nil, err
	}

	for key, val := range perms {
		var arr []string
		if err := json.Unmarshal(val, &arr); err != nil {
			continue // 配列でなければスキップ
		}
		sort.Strings(arr)
		sorted, err := json.Marshal(arr)
		if err != nil {
			return nil, err
		}
		perms[key] = sorted
	}

	return marshalRaw(perms)
}

// sortSandbox は sandbox オブジェクト内の allowedDomains をソートする｡
func sortSandbox(raw json.RawMessage) (json.RawMessage, error) {
	var sandbox map[string]json.RawMessage
	if err := json.Unmarshal(raw, &sandbox); err != nil {
		return nil, err
	}

	if networkRaw, ok := sandbox["network"]; ok {
		sorted, err := sortNetworkDomains(networkRaw)
		if err != nil {
			return nil, err
		}
		sandbox["network"] = sorted
	}

	return marshalRaw(sandbox)
}

// sortNetworkDomains は network オブジェクト内の allowedDomains をソートする｡
func sortNetworkDomains(raw json.RawMessage) (json.RawMessage, error) {
	var network map[string]json.RawMessage
	if err := json.Unmarshal(raw, &network); err != nil {
		return nil, err
	}

	if domainsRaw, ok := network["allowedDomains"]; ok {
		var domains []string
		if err := json.Unmarshal(domainsRaw, &domains); err != nil {
			return nil, err
		}
		sort.Strings(domains)
		sorted, err := json.Marshal(domains)
		if err != nil {
			return nil, err
		}
		network["allowedDomains"] = sorted
	}

	return marshalRaw(network)
}

// sortObjectKeys は JSON オブジェクトのキーをソートして再マーシャルする｡
func sortObjectKeys(raw json.RawMessage) (json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return marshalRaw(m)
}

// marshalRaw は map[string]json.RawMessage をキーソート済みの JSON に変換する｡
func marshalRaw(m map[string]json.RawMessage) (json.RawMessage, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// marshalIndent はトップレベル map を 2 スペースインデント + 末尾改行で出力する｡
func marshalIndent(m map[string]json.RawMessage) ([]byte, error) {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	// 末尾改行を付与
	b = append(b, '\n')
	return b, nil
}
