package vscode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ChangeTheme(newProfile string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not get os.UserConfigDir: %w", err)
	}

	vscodeSettings := filepath.Join(configDir, "Code", "User", "settings.json")
	content, err := os.ReadFile(vscodeSettings)
	if err != nil {
		return fmt.Errorf("could not read settings.json: %w", err)
	}

	var newJson map[string]interface{}
	err = json.Unmarshal(content, &newJson)
	if err != nil {
		return fmt.Errorf("could not unmarshal settings.json: %w", err)
	}

	newJson["workbench.colorTheme"] = newProfile
	newJsonString, err := json.MarshalIndent(newJson, "", "    ")
	if err != nil {
		return fmt.Errorf("could not marshal json to overwrite settings.json: %w", err)
	}

	// non-atomic write :S
	err = os.WriteFile(vscodeSettings, newJsonString, 0664)
	if err != nil {
		return fmt.Errorf("could not overwrite settings.json: %w", err)
	}

	return nil
}
