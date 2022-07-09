package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// naïve helper for a naïve script
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func vscode(newMode string) {
	configDir, err := os.UserConfigDir()
	check(err)
	vscodeSettings := filepath.Join(configDir, "Code", "User", "settings.json")
	content, err := os.ReadFile(vscodeSettings)
	check(err)

	var newJson map[string]interface{}
	err = json.Unmarshal(content, &newJson)
	check(err)
	// code --list-extensions --category themes
	// ^ gives us all but the builtin themes

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

func proc(command string, args ...string) (stdout string, err error) {
	cmd := exec.Command(command, args...)
	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	err = cmd.Start()
	if err != nil {
		return "", err
	}

	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	return stdoutBuf.String(), nil
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	// by convention, I'm switching between "dark" and "light"
	newMode := os.Args[1]

	fmt.Println("calle setting mode", newMode)

	vscode(newMode)
	fmt.Println("calle done setting mode", newMode)
}
