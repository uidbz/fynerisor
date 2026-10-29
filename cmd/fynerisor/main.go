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
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/parser"

	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/core/db"
	"github.com/uidbz/fynerisor/gui"
	guidb "github.com/uidbz/fynerisor/gui/db"
)

// Exit codes (issue #43): 0 success, 1 script error, 2 usage error.

func main() {
	os.Exit(dispatch(os.Args[1:]))
}

func dispatch(argv []string) int {
	if len(argv) > 0 {
		switch argv[0] {
		case "check":
			return runCheck(argv[1:])
		case "analyze":
			return runAnalyze(argv[1:])
		case "docs":
			return runDocs(argv[1:])
		case "run":
			return runRun(argv[1:])
		case "help", "-h", "--help":
			usage()
			return 0
		}
	}
	// Bare invocation stays the run command (issue #43 keeps the old
	// form as an alias).
	return runRun(argv)
}

func usage() {
	fmt.Fprintf(os.Stderr, "fynerisor - Execute Risor scripts with Fyne GUI\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  fynerisor [options] <script.risor> [script-args...]     run a script (GUI)\n")
	fmt.Fprintf(os.Stderr, "  fynerisor run   [options] <script.risor> [script-args...]  run; --core runs headless without Fyne\n")
	fmt.Fprintf(os.Stderr, "  fynerisor check <script.risor>...                       compile only; exit 1 on errors\n")
	fmt.Fprintf(os.Stderr, "  fynerisor analyze [--json] <script.risor>               requirements and imports\n")
	fmt.Fprintf(os.Stderr, "  fynerisor docs [--json] [-o file]                       generated API reference (issue #47)\n")
	fmt.Fprintf(os.Stderr, "  cat script.risor | fynerisor [options] -\n\n")
	fmt.Fprintf(os.Stderr, "Run options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExit codes: 0 success, 1 script error, 2 usage error.\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  fynerisor app.risor\n")
	fmt.Fprintf(os.Stderr, "  fynerisor --title \"My App\" --width 1024 --height 768 app.risor\n")
	fmt.Fprintf(os.Stderr, "  fynerisor --watch app.risor\n")
	fmt.Fprintf(os.Stderr, "  fynerisor --globals data.json app.risor arg1 arg2\n")
	fmt.Fprintf(os.Stderr, "  cat app.risor | fynerisor --headless -\n")
	fmt.Fprintf(os.Stderr, "  cat data.csv | fynerisor run --core -e 'print(json.marshal(csv.parse(io.stdin.read())))' | jq .\n")
	fmt.Fprintf(os.Stderr, "  fynerisor check app.risor lib.risor\n")
	fmt.Fprintf(os.Stderr, "  fynerisor analyze --json app.risor\n")
}

// ---- run ------------------------------------------------------------------

func runRun(argv []string) int {
	fs := flag.NewFlagSet("fynerisor", flag.ContinueOnError)
	fs.Usage = usage
	var (
		title     = fs.String("title", "", "Window title (default: script filename)")
		width     = fs.Int("width", 800, "Window width")
		height    = fs.Int("height", 600, "Window height")
		watch     = fs.Bool("watch", false, "Watch script file for changes and auto-reload")
		verbose   = fs.Bool("verbose", false, "Print verbose execution status")
		headless  = fs.Bool("headless", false, "Run in headless mode via the Fyne test driver (for testing GUI scripts)")
		coreMode  = fs.Bool("core", false, "Run headless through the core runtime: no Fyne at all (fails on GUI calls)")
		guiMode   = fs.Bool("gui", false, "Force the GUI runtime even for scripts that could run headless")
		globals   = fs.String("globals", "", "Path to JSON file with custom globals")
		execCode  = fs.String("e", "", "Execute inline code instead of a script file (stdin stays available to the script)")
		themeFlag = fs.String("theme", "", "Force theme: 'dark' or 'light' (default: system preference)")
		timeout   = fs.Duration("timeout", 0, "Abort the script after this duration (0: 30s for --headless, none for --core)")
		maxSteps  = fs.Int64("max-steps", 0, "Maximum VM steps (0: unlimited)")
		version   = fs.Bool("version", false, "Print version and exit")
	)
	if err := fs.Parse(argv); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if *version {
		fmt.Printf("fynerisor version %s\n", core.Version)
		return 0
	}

	args := fs.Args()
	if len(args) == 0 && *execCode == "" {
		usage()
		return 2
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
			fmt.Fprintf(os.Stderr, "fynerisor: reading stdin: %v\n", err)
			return 1
		}
		script = string(data)
		if *title == "" {
			*title = "fynerisor"
		}
	default:
		script, err = readScript(scriptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", scriptPath, err)
			return 1
		}
		if *title == "" {
			*title = filepath.Base(scriptPath)
		}
	}

	globalsMap := map[string]any{}
	if *globals != "" {
		globalsMap, err = loadGlobalsFile(*globals)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor: %v\n", err)
			return 2
		}
	}

	// The core (no-Fyne) runtime: --core, or --headless together with a
	// script that touches no GUI global (issue #43). Scripts that use
	// widget/window/canvas/... or require("@gui") keep the GUI runtime;
	// --gui pins it explicitly.
	usesGUI, err := scriptUsesGUI(script)
	if err != nil {
		fmt.Fprintln(os.Stderr, formatScriptError(scriptPath, err))
		return 1
	}
	if *coreMode && usesGUI {
		fmt.Fprintf(os.Stderr, "fynerisor: %s uses GUI globals; --core cannot run it (drop --core)\n", scriptPath)
		return 1
	}
	if *coreMode || (*headless && !usesGUI && !*guiMode) {
		return runCore(script, scriptPath, scriptArgs, globalsMap, *timeout, *maxSteps)
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
			fmt.Fprintf(os.Stderr, "fynerisor: invalid theme %q: must be 'dark' or 'light'\n", *themeFlag)
			return 2
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
		guidb.WithSQL(), // opt-in subpackage: links DB drivers only here (issue #40)
		gui.WithTie(),
		gui.WithIO(),
		gui.WithJSON(),
		gui.WithCSV(),
		gui.WithExec(),
		gui.WithScriptArgs(scriptArgs...),
		gui.WithGlobal("args", scriptArgs),
	)
	for name, value := range globalsMap {
		opts = append(opts, gui.WithGlobal(name, value))
	}
	// VM tuning goes through WithRisorOptions, its documented role
	// (issue #43: --timeout / --max-steps).
	var vmOpts []risor.Option
	if *timeout > 0 {
		vmOpts = append(vmOpts, risor.WithTimeout(*timeout))
	}
	if *maxSteps > 0 {
		vmOpts = append(vmOpts, risor.WithMaxSteps(*maxSteps))
	}
	if len(vmOpts) > 0 {
		opts = append(opts, gui.WithRisorOptions(vmOpts...))
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

	// In headless mode, run the script body to completion and exit with
	// its result: the typed lifecycle replaces polling the status
	// string, which a script could also set itself (issue #39).
	if *headless {
		loadScript(fyneWindow, script, scriptPath)
		limit := *timeout
		if limit == 0 {
			limit = 30 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), limit)
		err := fyneWindow.Run(ctx)
		cancel()
		switch {
		case err == nil:
			return 0
		case errors.Is(err, context.DeadlineExceeded):
			fmt.Fprintf(os.Stderr, "fynerisor: script execution timed out after %v\n", limit)
			return 1
		default:
			fmt.Fprintln(os.Stderr, formatScriptError(scriptPath, err))
			return 1
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
	return 0
}

// runCore executes the script through core.Context: no Fyne, process
// streams as io.stdin/stdout/stderr (issue #43).
func runCore(script, scriptPath string, scriptArgs []string, globalsMap map[string]any, timeout time.Duration, maxSteps int64) int {
	opts := []core.Option{
		core.WithHTTP(),
		core.WithOS(),
		core.WithStrings(),
		core.WithFilepath(),
		core.WithTime(),
		db.WithSQL(),
		core.WithTie(),
		core.WithIO(),
		core.WithJSON(),
		core.WithCSV(),
		core.WithExec(),
		core.WithScriptArgs(scriptArgs...),
		core.WithGlobal("args", scriptArgs),
	}
	for name, value := range globalsMap {
		opts = append(opts, core.WithGlobal(name, value))
	}
	if maxSteps > 0 {
		opts = append(opts, core.WithRisorOptions(risor.WithMaxSteps(maxSteps)))
	}
	cb := core.NewContext(opts...)

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	result, err := cb.EvalContext(ctx, script)
	if err != nil {
		fmt.Fprintln(os.Stderr, formatScriptError(scriptPath, err))
		return 1
	}
	if result != nil {
		fmt.Println(result)
	}
	return 0
}

// scriptUsesGUI reports whether the script touches GUI-only globals
// (or requires @gui). Conservative: parse failures report no GUI usage
// so the caller's compile check produces the better error.
func scriptUsesGUI(script string) (bool, error) {
	reqs, err := core.AnalyzeRequirements(script)
	if err != nil {
		return false, err
	}
	if reqs.RequiresGUI() {
		return true, nil
	}
	uses, err := core.UsesGlobals(script,
		"widget", "window", "canvas", "container", "chart", "dialog",
		"fyne", "clipboard", "constants", "binding", "go")
	if err != nil {
		return false, err
	}
	return uses, nil
}

// ---- check ----------------------------------------------------------------

func runCheck(argv []string) int {
	fs := flag.NewFlagSet("fynerisor check", flag.ContinueOnError)
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "usage: fynerisor check <script.risor>...")
		return 2
	}

	rc := 0
	for _, path := range files {
		var script string
		if path == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "fynerisor: reading stdin: %v\n", err)
				return 1
			}
			script = string(data)
		} else {
			s, err := readScript(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				rc = 1
				continue
			}
			script = s
		}
		cb := core.NewContext(
			core.WithHTTP(), core.WithOS(), core.WithStrings(), core.WithFilepath(),
			core.WithTime(), db.WithSQL(), core.WithTie(), core.WithIO(),
			core.WithJSON(), core.WithCSV(), core.WithExec(),
		)
		if err := cb.CompileScript(script); err != nil {
			fmt.Fprintln(os.Stderr, formatScriptError(path, err))
			rc = 1
		}
	}
	return rc
}

// ---- analyze --------------------------------------------------------------

func runAnalyze(argv []string) int {
	fs := flag.NewFlagSet("fynerisor analyze", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	files := fs.Args()
	if len(files) != 1 {
		fmt.Fprintln(os.Stderr, "usage: fynerisor analyze [--json] <script.risor>")
		return 2
	}

	path := files[0]
	var script string
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor: reading stdin: %v\n", err)
			return 1
		}
		script = string(data)
	} else {
		s, err := readScript(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			return 1
		}
		script = s
	}

	reqs, err := core.AnalyzeRequirements(script)
	if err != nil {
		fmt.Fprintln(os.Stderr, formatScriptError(path, err))
		return 1
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(reqs); err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Printf("%s\n", reqs.String())
	return 0
}

// ---- shared helpers --------------------------------------------------------

// locationRe extracts "location: file:line:col" from Risor compile
// errors so the CLI can re-emit the editor/CI-friendly form.
var locationRe = regexp.MustCompile(`location: (\S+?):(\d+):(\d+)`)

// formatScriptError renders a script error as path:line:col: message
// when position information is available (issue #43); plain
// "path: message" otherwise.
func formatScriptError(path string, err error) string {
	if pe := firstParserError(err); pe != nil {
		pos := pe.StartPosition()
		if pos.LineNumber() > 0 {
			return fmt.Sprintf("%s:%d:%d: %s: %s", path, pos.LineNumber(), pos.Column+1,
				strings.ToLower(pe.Type()), pe.Message())
		}
		return fmt.Sprintf("%s: %s: %s", path, strings.ToLower(pe.Type()), pe.Message())
	}
	text := strings.TrimSpace(err.Error())
	if m := locationRe.FindStringSubmatch(text); m != nil {
		first := text
		if i := strings.IndexByte(first, '\n'); i >= 0 {
			first = first[:i]
		}
		return fmt.Sprintf("%s:%s:%s: %s", path, m[2], m[3], first)
	}
	return fmt.Sprintf("%s: %v", path, err)
}

// firstParserError digs the first Risor parser error (with its
// position) out of the returned error, single or multi.
func firstParserError(err error) parser.ParserError {
	var errs *parser.Errors
	if errors.As(err, &errs) {
		if list := errs.Errors(); len(list) > 0 {
			return list[0]
		}
	}
	var se *parser.SyntaxError
	if errors.As(err, &se) {
		return se
	}
	var pe parser.ParserError
	if errors.As(err, &pe) {
		return pe
	}
	return nil
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
