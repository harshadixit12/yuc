package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/harshadixit12/yuc/config"
	"github.com/harshadixit12/yuc/output"
	"github.com/harshadixit12/yuc/scanner"
)

const version = "0.1.0"

const helpText = `yuc — Unicode hazard scanner for YAML files

USAGE:
  yuc [flags] <file> [file...]

FLAGS:
  -c, --config <path>    Path to a .yuc.yaml config file
      --no-color         Disable ANSI color output
      --list-categories  Print all built-in risk categories and exit
  -v, --version          Print version and exit
  -h, --help             Print this help and exit

EXIT CODES:
  0   Clean — no issues found
  1   Issues found (errors or warnings)
  2   Bad usage / unreadable file
  3   Config file parse error

EXAMPLES:
  yuc config.yaml
  yuc --config .yuc.yaml values.yaml
  yuc --list-categories

DOCUMENTATION:
  https://github.com/harshadixit12/yuc
`

func main() {
	os.Exit(Run())
}

func Run() int {
	fs := flag.NewFlagSet("yuc", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, helpText)
	}

	var (
		configPath     string
		noColor        bool
		listCategories bool
		showVersion    bool
	)

	fs.StringVar(&configPath, "config", "", "")
	fs.StringVar(&configPath, "c", "", "")
	fs.BoolVar(&noColor, "no-color", false, "")
	fs.BoolVar(&listCategories, "list-categories", false, "")
	fs.BoolVar(&showVersion, "version", false, "")
	fs.BoolVar(&showVersion, "v", false, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Fprint(os.Stdout, helpText)
			return 0
		}
		fmt.Fprintf(os.Stderr, "yuc: %v\n", err)
		return 2
	}

	// --version
	if showVersion {
		fmt.Printf("yuc %s\n", version)
		return 0
	}

	// Auto-disable color if not a TTY
	if !isTerminal(os.Stdout) {
		noColor = true
	}

	renderer := output.New(os.Stdout, output.Options{
		NoColor: noColor,
	})

	// --list-categories
	if listCategories {
		printCategoryList(noColor)
		return 0
	}

	// Load config
	var cfg config.Config
	var cfgErr error
	if configPath != "" {
		cfg, cfgErr = config.LoadFile(configPath)
		if cfgErr != nil {
			fmt.Fprintf(os.Stderr, "yuc: config error: %v\n", cfgErr)
			return 3
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// Collect files
	args := fs.Args()

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, helpText)
		return 2
	}

	var results []scanner.Result

	for _, path := range args {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "yuc: cannot open %q: %v\n", path, err)
			return 2
		}
		res, err := scanner.Scan(f, cfg, path)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "yuc: read error in %q: %v\n", path, err)
			return 2
		}
		results = append(results, res)
	}

	toRender := make([]scanner.Result, 0, len(results))

	for _, r := range results {
		if len(r.Issues) > 0 {
			toRender = append(toRender, r)
		}
	}

	renderer.RenderResults(toRender, results)

	// Exit code
	if output.HasErrors(results) || output.HasWarnings(results) {
		return 1
	}
	return 0
}

func printCategoryList(noColor bool) {
	cats := config.AllCategories()

	// Sort by name
	names := make([]string, 0, len(cats))
	for k := range cats {
		names = append(names, k)
	}
	sort.Strings(names)

	reset := "\033[0m"
	bold := "\033[1m"
	red := "\033[31m"
	yellow := "\033[33m"
	dimWhite := "\033[37m"
	cyan := "\033[36m"

	c := func(codes ...string) string {
		if noColor {
			return ""
		}
		return strings.Join(codes, "")
	}

	fmt.Println(c(bold) + "yuc built-in risk categories:" + c(reset))
	fmt.Println("")

	// Map category to its codepoint entries for display
	defaults := config.DefaultConfig()
	catChars := make(map[string][]config.CharEntry)
	for _, ch := range defaults.Chars {
		catChars[ch.Category] = append(catChars[ch.Category], ch)
	}

	for _, name := range names {
		cat := cats[name]
		sevColor := c(yellow)
		if cat.Severity == "error" {
			sevColor = c(red)
		}
		fmt.Printf("  %s%s%s\n", c(bold, cyan), name, c(reset))
		fmt.Printf("    severity: %s%s%s  enabled: %v\n",
			sevColor, cat.Severity, c(reset), cat.Enabled)
		chars := catChars[name]
		fmt.Printf("    %s%d codepoints%s\n", c(dimWhite), len(chars), c(reset))
		for _, ch := range chars {
			fmt.Printf("      U+%04X  %s\n", ch.Codepoint, ch.Name)
		}
		fmt.Println("")
	}
}

// isTerminal checks if the given file is a terminal (TTY).
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
