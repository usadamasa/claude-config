package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Settings は settings.json の構造を表す｡
type Settings struct {
	EnabledPlugins map[string]bool `json:"enabledPlugins"`
	Permissions    struct {
		Allow []string `json:"allow"`
		Deny  []string `json:"deny"`
		Ask   []string `json:"ask"`
	} `json:"permissions"`
	Sandbox struct {
		Network struct {
			AllowedDomains []string `json:"allowedDomains"`
		} `json:"network"`
	} `json:"sandbox"`
}

// CountEnabledPlugins は有効なプラグイン数を返す｡
func (s *Settings) CountEnabledPlugins() int {
	count := 0
	for _, enabled := range s.EnabledPlugins {
		if enabled {
			count++
		}
	}
	return count
}

// ListEnabledPluginNames は有効なプラグイン名の一覧をソート済みで返す｡
func (s *Settings) ListEnabledPluginNames() []string {
	var names []string
	for name, enabled := range s.EnabledPlugins {
		if enabled {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// Load は settings.json を読み込んでパースする｡
func Load(path string) (*Settings, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- CLIツール: パスはフラグ引数由来
	if err != nil {
		return nil, err
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("settings JSON のパースに失敗: %w", err)
	}

	return &s, nil
}
