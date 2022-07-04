package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func createFolder(foldername string) error {
	_, err := os.Stat(foldername)
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(foldername, 0755)
		if err != nil {
			return fmt.Errorf("could not create folder %s: %w", foldername, err)
		}
	} else if err != nil {
		return fmt.Errorf("could not find folder %s: %w", foldername, err)
	}
	return nil
}

func fileWithDefaultContent(filename, defaultContent string) (string, error) {
	content, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("file %s not found, trying to create it\n", filename)
		err = os.WriteFile(filename, []byte(defaultContent), 0664)
		if err != nil {
			return "", fmt.Errorf("could not write file %s: %w", filename, err)
		} else {
			return defaultContent, nil
		}
	} else if err != nil {
		return "", fmt.Errorf("could not stat file %s: %w", filename, err)
	}
	return string(content), nil
}

type config struct {
	wanted string
	modes  []string
}

// This will set a default config (touch the file system) if no config exists.
func getConfig() (*config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user config dir: %w", err)
	}

	slDir := filepath.Join(configDir, "signal-lamp")
	err = createFolder(slDir)
	if err != nil {
		return nil, err
	}
	err = createFolder(filepath.Join(slDir, "triggers"))
	if err != nil {
		return nil, err
	}

	modes, err := fileWithDefaultContent(filepath.Join(slDir, "modes"), "dark\nlight\n")
	if err != nil {
		return nil, err
	}

	wanted, err := fileWithDefaultContent(filepath.Join(slDir, "wanted"), "dark\n")
	if err != nil {
		return nil, err
	}

	return &config{
		wanted: strings.TrimSpace(wanted),
		modes:  strings.Split(strings.TrimSpace(modes), "\n"),
	}, nil
}

func nextMode(current string, available []string) string {
	for i, v := range available {
		if current == v {
			return available[(i+1)%len(available)]
		}
	}
	// current is not in available, which is weird.
	fmt.Println("Oops, did not find current mode", current, "in available modes", available)
	return ""
}

func triggerScripts(newMode string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	triggers := filepath.Join(configDir, "signal-lamp", "triggers")
	matches, err := filepath.Glob(triggers + "/*")
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, m := range matches {
		wg.Add(1)
		go func(script string) {
			fmt.Println("running script", script)
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(script, newMode)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Start(); err != nil {
				fmt.Println("error 1")
			}
			if err := cmd.Wait(); err != nil {
				fmt.Println("error 2")
			}
			wg.Done()
		}(m)
	}
	wg.Wait()
	return nil
}

func saveMode(mode string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	wanted := filepath.Join(configDir, "signal-lamp", "wanted")
	err = os.WriteFile(wanted, []byte(mode), 0664)
	return err
}

func main() {
	var (
		readConf = flag.Bool("q", false, "read config without changing mode")
		toggle   = flag.Bool("t", false, "toggle config")
	)
	flag.Parse()

	config, err := getConfig()
	if err != nil {
		panic(err)
	}
	currentMode, availableModes := config.wanted, config.modes

	if *readConf {
		fmt.Println(currentMode)
		os.Exit(0)
	}

	if *toggle {
		newMode := nextMode(currentMode, availableModes)
		if newMode != currentMode {
			fmt.Println("new mode", newMode)

			err = triggerScripts(newMode)
			if err != nil {
				panic(err)
			}

			err = saveMode(newMode)
			if err != nil {
				panic(err)
			}
		}
	}
}
