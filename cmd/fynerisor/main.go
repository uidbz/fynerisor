package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/fsnotify/fsnotify"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/gui"
)

var (
	title      = flag.String("title", "", "Window title (default: script filename)")
	width      = flag.Int("width", 800, "Window width")
	height     = flag.Int("height", 600, "Window height")
	watch      = flag.Bool("watch", false, "Watch script file for changes and auto-reload")
	verbose    = flag.Bool("verbose", false, "Print verbose execution status")
	headless   = flag.Bool("headless", false, "Run in headless mode (for testing)")
	globals    = flag.String("globals", "", "Path to JSON file with custom globals")
	themeFlag  = flag.String("theme", "", "Force theme: 'dark' or 'light' (default: system preference)")
	versionPtr = flag.Bool("version", false, "Print version and exit")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "fynerisor - Execute Risor scripts with Fyne GUI\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  fynerisor [options] <script.risor> [script-args...]\n")
		fmt.Fprintf(os.Stderr, "  cat script.risor | fynerisor [options] -\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  fynerisor app.risor\n")
		fmt.Fprintf(os.Stderr, "  fynerisor --title \"My App\" --width 1024 --height 768 app.risor\n")
		fmt.Fprintf(os.Stderr, "  fynerisor --watch app.risor\n")
		fmt.Fprintf(os.Stderr, "  fynerisor --globals data.json app.risor arg1 arg2\n")
		fmt.Fprintf(os.Stderr, "  cat app.risor | fynerisor --headless -\n")
	}

	flag.Parse()

	if *versionPtr {
		fmt.Printf("fynerisor version %s\n", core.Version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	scriptPath := args[0]
	scriptArgs := args[1:]

	// Read script
	var script string
	var err error
	if scriptPath == "-" {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		script = string(data)
		if *title == "" {
			*title = "fynerisor"
		}
	} else {
		script, err = readScript(scriptPath)
		if err != nil {
			log.Fatalf("Error reading script: %v", err)
		}
		if *title == "" {
			*title = filepath.Base(scriptPath)
		}
	}

	// Load custom globals if specified
	var customGlobals []risor.Option
	if *globals != "" {
		customGlobals, err = loadGlobalsFile(*globals)
		if err != nil {
			log.Fatalf("Error loading globals file: %v", err)
		}
	}

	// Add script arguments as global
	scriptArgsGlobal := map[string]any{
		"args": scriptArgs,
	}
	customGlobals = append(customGlobals, risor.WithEnv(scriptArgsGlobal))

	// Create app and window
	var a fyne.App
	var w fyne.Window

	if *headless {
		a = test.NewApp()
		w = a.NewWindow(*title)
	} else {
		a = app.New()

		// Set theme if specified
		switch strings.ToLower(*themeFlag) {
		case "dark":
			a.Settings().SetTheme(theme.DarkTheme())
		case "light":
			a.Settings().SetTheme(theme.LightTheme())
		case "":
			// Use default (system preference)
		default:
			log.Fatalf("Invalid theme '%s': must be 'dark' or 'light'", *themeFlag)
		}

		w = a.NewWindow(*title)
		w.Resize(fyne.NewSize(float32(*width), float32(*height)))
	}

	// Create fynerisor window with standard modules enabled
	var fyneWindow *gui.Window
	if *verbose {
		fyneWindow = gui.NewWindow(w,
			gui.WithHTTP(),
			gui.WithOS(),
			gui.WithStrings(),
			gui.WithFilepath(),
			gui.WithTime(),
			gui.WithSQL(),
			gui.WithIO(),
			gui.WithRisorOptions(customGlobals...),
			gui.WithStatusCallback(func(status string) {
				log.Printf("[STATUS] %s", status)
			}),
			gui.WithResultCallback(func(result string) {
				log.Printf("[RESULT] %s", result)
			}),
		)
	} else {
		fyneWindow = gui.NewWindow(w,
			gui.WithHTTP(),
			gui.WithOS(),
			gui.WithStrings(),
			gui.WithFilepath(),
			gui.WithTime(),
			gui.WithSQL(),
			gui.WithIO(),
			gui.WithRisorOptions(customGlobals...),
			gui.WithStatusCallback(func(status string) {
				if strings.HasPrefix(status, "ERROR:") {
					log.Printf("[ERROR] %s", status)
				}
			}),
		)
	}

	// Analyze script requirements and handle imports
	reqs, err := core.AnalyzeRequirements(script)
	if err != nil {
		log.Fatalf("Error analyzing script: %v", err)
	}

	// Import dependencies in order
	for _, importPath := range reqs.Imports {
		if *verbose {
			log.Printf("Importing: %s", importPath)
		}
		err := fyneWindow.ImportScript(importPath)
		if err != nil {
			log.Fatalf("Error importing %s: %v", importPath, err)
		}
	}

	// Execute script
	executeScript(fyneWindow, script)

	// Watch mode
	if *watch && scriptPath != "-" {
		go watchScript(scriptPath, fyneWindow)
	}

	// In headless mode, wait for script to complete
	if *headless {
		// Wait up to 30 seconds for script to complete
		for i := 0; i < 300; i++ {
			time.Sleep(100 * time.Millisecond)
			if fyneWindow.Status == "Ready!" {
				os.Exit(0)
			}
			if strings.HasPrefix(fyneWindow.Status, "ERROR:") {
				log.Fatalf("Script execution failed: %s", fyneWindow.Status)
			}
		}
		log.Fatalf("Script execution timed out after 30 seconds")
	}

	// Show and run
	fyneWindow.ShowAndRun()
}

func readScript(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

func executeScript(fw *gui.Window, script string) {
	fw.LoadScript(script)
	fw.Execute()
}

func watchScript(path string, fw *gui.Window) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		log.Printf("Failed to watch file: %v", err)
		return
	}

	log.Printf("Watching %s for changes...", path)

	// Debounce timer to avoid multiple reloads
	var debounceTimer *time.Timer
	debounce := 100 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Cancel any pending reload
				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				// Schedule reload after debounce period
				debounceTimer = time.AfterFunc(debounce, func() {
					log.Printf("File modified, reloading...")
					script, err := readScript(path)
					if err != nil {
						log.Printf("Error reading script: %v", err)
						return
					}

					// Execute in main thread
					fyne.Do(func() {
						executeScript(fw, script)
					})
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func loadGlobalsFile(path string) ([]risor.Option, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read globals file: %w", err)
	}

	var globalsMap map[string]any
	if err := json.Unmarshal(data, &globalsMap); err != nil {
		return nil, fmt.Errorf("parse globals JSON: %w", err)
	}

	return []risor.Option{risor.WithEnv(globalsMap)}, nil
}
