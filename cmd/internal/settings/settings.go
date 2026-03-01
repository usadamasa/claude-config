package settings

import (
	"encoding/json"
	"fmt"
	"os"
)

// Settings は settings.json の構造を表す｡
type Settings struct {
	Permissions struct {
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
