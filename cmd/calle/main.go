package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func vscode(configDir, newMode string) {
	vscodeSettings := filepath.Join(configDir, "Code", "User", "settings.json")
	content, err := os.ReadFile(vscodeSettings)
	check(err)

	var newJson map[string]interface{}
	err = json.Unmarshal(content, &newJson)
	check(err)

	const vscodeLight = "Default Light+"
	const vscodeDark = "Arc Dark"
	if newMode == "dark" {
		newJson["workbench.colorTheme"] = vscodeDark
	} else {
		newJson["workbench.colorTheme"] = vscodeLight
	}
	newJsonString, err := json.MarshalIndent(newJson, "", "    ")
	check(err)

	// non-atomic write :S
	err = os.WriteFile(vscodeSettings, newJsonString, 0664)
	check(err)
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	newMode := os.Args[1]

	fmt.Println("calle setting mode", newMode)

	configDir, err := os.UserConfigDir()
	check(err)

	vscode(configDir, newMode)
}
