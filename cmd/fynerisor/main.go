package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	execCode   = flag.String("e", "", "Execute inline code instead of a script file (stdin stays available to the script)")
	themeFlag  = flag.String("theme", "", "Force theme: 'dark' or 'light' (default: system preference)")
	versionPtr = flag.Bool("version", false, "Print version and exit")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "fynerisor - Execute Risor scripts with Fyne GUI\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  fynerisor [options] <script.risor> [script-args...]\n")
		fmt.Fprintf(os.Stderr, "  fynerisor [options] -e '<code>' [script-args...]\n")
		fmt.Fprintf(os.Stderr, "  cat script.risor | fynerisor [options] -\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  fynerisor app.risor\n")
		fmt.Fprintf(os.Stderr, "  fynerisor --title \"My App\" --width 1024 --height 768 app.risor\n")
		fmt.Fprintf(os.Stderr, "  fynerisor --watch app.risor\n")
		fmt.Fprintf(os.Stderr, "  fynerisor --globals data.json app.risor arg1 arg2\n")
		fmt.Fprintf(os.Stderr, "  cat app.risor | fynerisor --headless -\n")
		fmt.Fprintf(os.Stderr, "  cat data.csv | fynerisor --headless -e 'print(io.stdin.read())'   # pipeline filter (issue #42)\n")
	}

	flag.Parse()

	if *versionPtr {
		fmt.Printf("fynerisor version %s\n", core.Version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 && *execCode == "" {
		flag.Usage()
		os.Exit(1)
	}

	var scriptPath string
	var scriptArgs []string
	if *execCode == "" {
		scriptPath = args[0]
		scriptArgs = args[1:]
	} else {
		// Inline code: every positional arg is a script arg (issue #42).
		scriptArgs = args
	}

	// Read script
	var script string
	var err error
	switch {
	case *execCode != "":
		// Inline code; stdin stays a data stream for the script (issue #42).
		script = *execCode
		if *title == "" {
			*title = "fynerisor"
		}
	case scriptPath == "-":
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		script = string(data)
		if *title == "" {
			*title = "fynerisor"
		}
	default:
		script, err = readScript(scriptPath)
		if err != nil {
			log.Fatalf("Error reading script: %v", err)
		}
		if *title == "" {
			*title = filepath.Base(scriptPath)
		}
	}

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

	// Create fynerisor window with standard modules enabled. Named
	// globals (the args convenience variable, --globals JSON) go in per
	// key via WithGlobal: they then live in the shared environment that
	// imported modules are compiled and run against too, instead of
	// being opaque WithRisorOptions that modules cannot see (issue #27).
	var opts []gui.Option
	opts = append(opts,
		gui.WithHTTP(),
		gui.WithOS(),
		gui.WithStrings(),
		gui.WithFilepath(),
		gui.WithTime(),
		gui.WithSQL(),
		gui.WithTie(),
		gui.WithIO(),
		gui.WithJSON(),
		gui.WithCSV(),
		gui.WithScriptArgs(scriptArgs...),
		gui.WithGlobal("args", scriptArgs),
	)
	if *globals != "" {
		globalsMap, err := loadGlobalsFile(*globals)
		if err != nil {
			log.Fatalf("Error loading globals file: %v", err)
		}
		for name, value := range globalsMap {
			opts = append(opts, gui.WithGlobal(name, value))
		}
	}
	if *verbose {
		opts = append(opts,
			gui.WithStatusCallback(func(status string) {
				log.Printf("[STATUS] %s", status)
			}),
			gui.WithResultCallback(func(result string) {
				log.Printf("[RESULT] %s", result)
			}),
		)
	} else {
		opts = append(opts, gui.WithOnStateChange(func(state gui.RunState, err error) {
			if state == gui.StateFailed {
				log.Printf("[ERROR] %v", err)
			}
		}))
	}
	fyneWindow := gui.NewWindow(w, opts...)

	// Analyze script requirements (informational). Imports are resolved by
	// the runtime's import() builtin during execution; the removed
	// ImportScript API must not be called here (issue #22).
	reqs, err := core.AnalyzeRequirements(script)
	if err != nil {
		log.Fatalf("Error analyzing script: %v", err)
	}
	if *verbose {
		for _, importPath := range reqs.Imports {
			log.Printf("Analysed import (resolved at runtime): %s", importPath)
		}
	}

	// In headless mode, run the script body to completion and exit with its
	// result: the typed lifecycle replaces polling the status string, which
	// a script could also set itself (issue #39).
	if *headless {
		loadScript(fyneWindow, script, scriptPath)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := fyneWindow.Run(ctx)
		cancel()
		switch {
		case err == nil:
			os.Exit(0)
		case errors.Is(err, context.DeadlineExceeded):
			log.Fatalf("Script execution timed out after 30 seconds")
		default:
			log.Fatalf("Script execution failed: %v", err)
		}
	}

	// Execute script
	executeScript(fyneWindow, script, scriptPath)

	// Watch mode
	if *watch && scriptPath != "" && scriptPath != "-" {
		go watchScript(scriptPath, fyneWindow)
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

func executeScript(fw *gui.Window, script, scriptPath string) {
	loadScript(fw, script, scriptPath)
	fw.Execute()
}

func loadScript(fw *gui.Window, script, scriptPath string) {
	if scriptPath != "" && scriptPath != "-" {
		// Relative imports resolve against the script's own directory
		// (issue #8). For stdin ("-") keep the process CWD.
		fw.LoadScriptFrom(script, scriptPath)
	} else {
		fw.LoadScript(script)
	}
}

func watchScript(path string, fw *gui.Window) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	dir := filepath.Dir(path)
	if err := watcher.Add(dir); err != nil {
		log.Printf("Failed to watch %s: %v", dir, err)
		return
	}

	log.Printf("Watching %s for changes...", path)

	// Debounce timer to avoid multiple reloads
	var debounceTimer *time.Timer
	debounce := 250 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// The *directory* is watched, not the file: atomic saves
			// (vim, IDEs: write temp, rename/remove over the target)
			// detach a file watch after the first save (issue #29).
			if filepath.Base(event.Name) != filepath.Base(path) {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			// Cancel any pending reload
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			// Schedule reload after debounce period
			debounceTimer = time.AfterFunc(debounce, func() {
				script, err := readScript(path)
				if err != nil {
					// The editor may still be mid-save; a follow-up
					// event retries.
					log.Printf("Error reading script: %v", err)
					return
				}
				log.Printf("File modified, reloading...")

				// Execute in main thread; Execute stops the run it
				// replaces, so saves never stack VMs.
				fyne.Do(func() {
					executeScript(fw, script, path)
				})
			})
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func loadGlobalsFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read globals file: %w", err)
	}

	var globalsMap map[string]any
	if err := json.Unmarshal(data, &globalsMap); err != nil {
		return nil, fmt.Errorf("parse globals JSON: %w", err)
	}

	return globalsMap, nil
}
