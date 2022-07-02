package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// [theme-name] => uuid
func dconfDumpProfiles(dumpOutput string) (string, map[string]string, error) {
	profileUuidByName := make(map[string]string)
	profileBlocks := strings.Split(strings.TrimSpace(dumpOutput), "\n\n")
	// the first part is the meta block
	// alternatives : fmt.Sscanf, strings.Cut, strings.Split, regexp
	defaultLine := strings.Split(profileBlocks[0], "\n")[1]
	currentProfile := strings.Split(defaultLine, "'")[1]

	for _, p := range profileBlocks[1:] {
		lines := strings.Split(p, "\n")
		var uuid string

		_, err := fmt.Sscanf(lines[0], "[:%36s]", &uuid)
		if err != nil {
			return "", nil, fmt.Errorf("couldn't parse uuid from profile: %w", err)
		}
		fmt.Println("got uuid", uuid)

		// above workaround of specifying length does not work here, profile names are of variable length
		// the line we're trying to match: visible-name='dark'
		lastLinesParts := strings.Split(lines[len(lines)-1], "'")
		// in my testing, the profile name is always the last line. don't assume, check:
		if lastLinesParts[0] != "visible-name=" {
			return "", nil, fmt.Errorf("expected visible-name, got %v", lines)
		}
		name := lastLinesParts[1]
		fmt.Println("got profile name", name, "from last line", lines[len(lines)-1])

		profileUuidByName[name] = uuid
	}
	return currentProfile, profileUuidByName, nil
}

func gnomeTerminal(configDir, newMode string) {
	// ~/.config/dconf/user is a binary file that contains the name of my dark & light profiles
	// my profiles are "light" and "dark"
	// gnome-terminal --profile=dark works (hard to control though, would be nicer to set the default for new terminals, and change the currently open ones)
	// gsettings get org.gnome.Terminal.ProfilesList list is interesting
	// dconf dump /org/gnome/terminal/ is interesting
	// gsettings set org.gnome.Terminal.ProfilesList default <uuid> works, for new terminals
	// xdotool key --clearmodifiers Shift+F10 r 1 (or 2, 3, ...) works when a terminal is focused

	// list profiles, look for ones named ".*dark.*" or ".*light.*"
	// current window (to refocus): xdotool getwindowfocus

	_, err := proc("dconf", "dump", "/org/gnome/terminal/legacy/profiles:/")
	//stdout, err := proc("dconf", "dump", "/org/gnome/terminal/legacy/profiles:/")
	if err != nil {
		return
	}

	//profiles := dconfDumpProfiles(stdout)

}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	// by convention, I'm switching between "dark" and "light"
	newMode := os.Args[1]

	fmt.Println("calle setting mode", newMode)

	configDir, err := os.UserConfigDir()
	check(err)

	vscode(configDir, newMode)
	gnomeTerminal(configDir, newMode)
}
