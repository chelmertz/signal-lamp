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
	currentTheme string
	order        []string
	// key = theme name
	availableThemes map[string]themes
}

func (c *config) setTheme(theme string) error {
	for _, v := range c.order {
		if theme == v {
			c.currentTheme = theme
			return nil
		}
	}

	return fmt.Errorf("could not set theme to %s, it's not configured", theme)
}

func (c *config) next() error {
	for i, v := range c.order {
		if c.currentTheme == v {
			c.currentTheme = c.order[(i+1)%len(c.order)]
			return nil
		}
	}

	if len(c.order) > 0 {
		// this is auto-correcting:
		// if wanted was a wrongly spelled item of something in order, next() will be correct from now on
		c.currentTheme = c.order[0]
		return nil
	}

	return errors.New("could not find the next() of an empty order")
}

func (c *config) availableThemesString() string {
	var keys []string
	for m := range c.availableThemes {
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

	themesString, err := fileWithDefaultContent(filepath.Join(slDir, "signal-lamp.toml"), "[dark]\n[light]\n")
	if err != nil {
		return nil, err
	}
	var availableThemes map[string]themes
	_, err = toml.Decode(themesString, &availableThemes)
	if err != nil {
		return nil, fmt.Errorf("invalid toml: %w", err)
	}

	order := make([]string, 0)
	for key := range availableThemes {
		order = append(order, key)
	}

	wanted, err := fileWithDefaultContent(filepath.Join(slDir, "wanted"), "\n")
	if err != nil {
		return nil, err
	}

	return &config{
		currentTheme:    strings.TrimSpace(wanted),
		availableThemes: availableThemes,
		order:           order,
	}, nil
}

func changeThemes(themes themes) {
	var wg sync.WaitGroup

	callbacks := map[string]func(string) error{
		"gnometerminal": gnometerminal.ChangeProfile,
		"vscode":        vscode.ChangeTheme,
	}

	for cbName, cb := range callbacks {
		if themeName, ok := themes[cbName]; ok {
			wg.Add(1)
			go func(themeName_ string, cb_ func(string) error, cbName_ string) {
				err := cb_(themeName_)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error in %s: %s\n", cbName_, err)
				}
				wg.Done()
			}(themeName, cb, cbName)
		}
	}

	wg.Wait()
}

func saveTheme(theme string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	wanted := filepath.Join(configDir, "signal-lamp", "wanted")
	err = os.WriteFile(wanted, []byte(theme), 0664)
	if err != nil {
		return fmt.Errorf("could not save wanted theme: %w", err)
	}
	return nil
}

func main() {
	var (
		queryConfig = flag.Bool("query", false, "query current theme, without changing it")
		cycle       = flag.Bool("cycle", false, "cycle configured themes")
		theme       = flag.String("theme", "", "change theme")
	)
	flag.Parse()

	config, err := readAndSetDefaultConfig()
	if err != nil {
		panic(err)
	}

	if *queryConfig {
		fmt.Println("current theme:", config.currentTheme)
		fmt.Println("available themes:", config.availableThemesString())
		os.Exit(0)
	}

	if *theme != "" {
		err := config.setTheme(*theme)
		if err != nil {
			log.Fatal(err)
		}
	} else if *cycle {
		err := config.next()
		if err != nil {
			log.Fatal(err)
		}
	}

	changeThemes(config.availableThemes[config.currentTheme])

	err = saveTheme(config.currentTheme)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("current theme:", config.currentTheme)
}
