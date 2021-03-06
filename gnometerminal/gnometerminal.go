// Requires these CLI tools:
//   - wmctrl (list X windows)
//   - xdotool (act on X windows)
//   - xdotool (act on X windows)
//   - dconf (dump gnome terminal settings)
//   - gsettings (set gnome terminal settings)
package gnometerminal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Warning: the code in this function relies a lot on strings that happened to
// be in the output on my Ubuntu instance at the time of writing this code.
func ChangeProfile(newProfile string) error {
	// ~/.config/dconf/user is a binary file that contains the name of my dark & light profiles
	// my profiles are "light" and "dark"
	// gnome-terminal --profile=dark works (hard to control though, would be nicer to set the default for new terminals, and change the currently open ones)
	// gsettings get org.gnome.Terminal.ProfilesList list is interesting
	// dconf dump /org/gnome/terminal/ is interesting
	// gsettings set org.gnome.Terminal.ProfilesList default <uuid> works, for new terminals
	// xdotool key --clearmodifiers Shift+F10 r 1 (or 2, 3, ...) works when a terminal is focused

	// list profiles, look for ones named ".*dark.*" or ".*light.*"
	// current window (to refocus): xdotool getwindowfocus

	stdout, err := proc("dconf", "dump", "/org/gnome/terminal/legacy/profiles:/")
	if err != nil {
		return fmt.Errorf("could not execute dconf dump: %w", err)
	}

	dconfDump, err := dconfDumpProfiles(stdout)
	if err != nil {
		return fmt.Errorf("could not parse dconf dump: %w", err)
	}

	var newProfileId string
	for profileName, uuid := range dconfDump.profiles {
		if profileName == newProfile {
			newProfileId = uuid
			break
		}
	}

	if newProfileId == "" || newProfileId == dconfDump.current {
		fmt.Fprintln(os.Stderr, "not switching gnome terminal profile")
		return nil
	}

	// set default profile for new instances of gnome terminal
	_, err = proc("gsettings", "set", "org.gnome.Terminal.ProfilesList", "default", newProfileId)
	if err != nil {
		return fmt.Errorf("could not save with gsettings: %w", err)
	}

	// list currently opened gnome terminal window IDs
	xwindows, err := proc("wmctrl", "-lx")
	if err != nil {
		return fmt.Errorf("could not list windows with wmctrl: %w", err)
	}
	gnomeTerminalXWindowIds := make([]string, 0)
	for _, line := range strings.Split(xwindows, "\n") {
		if strings.Contains(line, "gnome-terminal-server.Gnome-terminal") {
			gnomeTerminalXWindowIds = append(gnomeTerminalXWindowIds, strings.Split(line, " ")[0])
		}
	}

	if len(gnomeTerminalXWindowIds) == 0 {
		return nil
	}

	// get window ID of currently active window, to be able to focus it later
	currentlyActiveXWindowId, err := proc("xdotool", "getwindowfocus")
	if err != nil {
		return err
	}

	newProfileIndex, err := gnomeTerminalProfileHotkey(dconfDump.order, newProfileId)
	if err != nil {
		return fmt.Errorf("could not find gnome terminal profile hotkey: %w", err)
	}

	// loop through open gnome terminal instances
	// set the current profile
	for _, windowId := range gnomeTerminalXWindowIds {
		// focus window

		// spent a lot of time going down the "wmctrl -ai windowId" road here, don't do that
		// WARNING: this relies on the "Enable the menu accelerator key (F10 by default)"
		// setting being active, you find it in Preferences > Global > General
		_, err = proc("xdotool", "windowfocus", "--sync", windowId, "key", "--clearmodifiers", "Shift+F10", "r", fmt.Sprint(newProfileIndex))
		if err != nil {
			return fmt.Errorf("could not set profile for terminal window %s with xdotool: %w", windowId, err)
		}
	}

	// focus the previously active window again
	_, err = proc("xdotool", "windowfocus", "--sync", currentlyActiveXWindowId)
	if err != nil {
		return fmt.Errorf("could not refocus previous application with xdotool: %w", err)
	}

	return nil
}

type dconfProfileDump struct {
	current string
	// [theme-name] => uuid
	profiles map[string]string
	order    []string
}

// we need the order of the profiles as well, to target the correct profile
// when changing the profile of the currently open gnome terminal windows
func dconfDumpProfiles(dumpOutput string) (*dconfProfileDump, error) {
	profileUuidByName := make(map[string]string)
	profileBlocks := strings.Split(strings.TrimSpace(dumpOutput), "\n\n")
	// the first part is the meta block
	// alternatives : fmt.Sscanf, strings.Cut, strings.Split, regexp
	metaLines := strings.Split(profileBlocks[0], "\n")
	currentProfile := strings.Split(metaLines[1], "'")[1]
	order := make([]string, 0)
	for _, commas := range strings.Split(metaLines[2], ",") {
		order = append(order, strings.Split(commas, "'")[1])
	}

	for _, p := range profileBlocks[1:] {
		lines := strings.Split(p, "\n")
		var uuid string

		_, err := fmt.Sscanf(lines[0], "[:%36s]", &uuid)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse uuid from profile: %w", err)
		}

		// above workaround of specifying length does not work here, profile names are of variable length
		// the line we're trying to match: visible-name='dark'
		lastLinesParts := strings.Split(lines[len(lines)-1], "'")
		// in my testing, the profile name is always the last line. don't assume, check:
		if lastLinesParts[0] != "visible-name=" {
			return nil, fmt.Errorf("expected visible-name, got %v", lines)
		}
		name := lastLinesParts[1]

		profileUuidByName[name] = uuid
	}

	return &dconfProfileDump{
		current:  currentProfile,
		profiles: profileUuidByName,
		order:    order,
	}, nil
}

func gnomeTerminalProfileHotkey(order []string, needle string) (int, error) {
	for i, uuid := range order {
		if uuid == needle {
			// gnome terminal profile menu is 1-indexed
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("could not find gnome terminal profile hotkey")
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
