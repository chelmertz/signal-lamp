package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/chelmertz/signal-lamp/gnometerminal"
	"github.com/chelmertz/signal-lamp/vscode"
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
		return "", fmt.Errorf("could not read file %s: %w", filename, err)
	}
	return string(content), nil
}

type themes map[string]string

type config struct {
	currentMode string
	order       []string
	// key = theme name
	availableModes map[string]themes
}

func (c *config) setMode(mode string) error {
	for _, v := range c.order {
		if mode == v {
			c.currentMode = mode
			return nil
		}
	}

	return fmt.Errorf("could not set mode to %s, it's not configured", mode)
}

func (c *config) next() error {
	for i, v := range c.order {
		if c.currentMode == v {
			c.currentMode = c.order[(i+1)%len(c.order)]
			return nil
		}
	}

	if len(c.order) > 0 {
		// this is auto-correcting:
		// if wanted was a wrongly spelled item of something in order, next() will be correct from now on
		c.currentMode = c.order[0]
		return nil
	}

	return errors.New("could not find the next() of an empty order")
}

func (c *config) availableModesString() string {
	var keys []string
	for m := range c.availableModes {
		keys = append(keys, m)
	}
	return strings.Join(keys, ", ")
}

// This will set a default config (touch the file system) if no config exists.
func readAndSetDefaultConfig() (*config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user config dir: %w", err)
	}

	slDir := filepath.Join(configDir, "signal-lamp")
	err = createFolder(slDir)
	if err != nil {
		return nil, err
	}

	modesString, err := fileWithDefaultContent(filepath.Join(slDir, "signal-lamp.toml"), "[dark]\n[light]\n")
	if err != nil {
		return nil, err
	}
	var modes map[string]themes
	_, err = toml.Decode(modesString, &modes)
	if err != nil {
		return nil, fmt.Errorf("invalid toml: %w", err)
	}

	order := make([]string, 0)
	for key := range modes {
		order = append(order, key)
	}

	wanted, err := fileWithDefaultContent(filepath.Join(slDir, "wanted"), "\n")
	if err != nil {
		return nil, err
	}

	return &config{
		currentMode:    strings.TrimSpace(wanted),
		availableModes: modes,
		order:          order,
	}, nil
}

func changeThemes(themes themes) {
	var wg sync.WaitGroup

	if name, ok := themes["gnometerminal"]; ok {
		wg.Add(1)
		go func(name string) {
			err := gnometerminal.ChangeProfile(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error in gnometerminal: %s\n", err)
			}
			wg.Done()
		}(name)
	}

	if name, ok := themes["vscode"]; ok {
		wg.Add(1)
		go func(name string) {
			err := vscode.ChangeTheme(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error in vscode: %s\n", err)
			}
			wg.Done()
		}(name)
	}

	wg.Wait()
}

func saveMode(mode string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	wanted := filepath.Join(configDir, "signal-lamp", "wanted")
	err = os.WriteFile(wanted, []byte(mode), 0664)
	if err != nil {
		return fmt.Errorf("could not save wanted mode: %w", err)
	}
	return nil
}

func main() {
	var (
		queryConfig = flag.Bool("query", false, "query current mode, without changing it")
		toggle      = flag.Bool("toggle", false, "toggle config")
		theme       = flag.String("theme", "", "change theme")
	)
	flag.Parse()

	config, err := readAndSetDefaultConfig()
	if err != nil {
		panic(err)
	}

	if *queryConfig {
		fmt.Println("current mode:", config.currentMode)
		fmt.Println("available modes:", config.availableModesString())
		os.Exit(0)
	}

	if *theme != "" {
		err := config.setMode(*theme)
		if err != nil {
			log.Fatal(err)
		}
	} else if *toggle {
		err := config.next()
		if err != nil {
			log.Fatal(err)
		}
	}

	changeThemes(config.availableModes[config.currentMode])

	err = saveMode(config.currentMode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("current mode:", config.currentMode)
}
