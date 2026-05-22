package output

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/harshadixit12/yuc/scanner"
)

// ANSI color codes
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	red       = "\033[31m"
	yellow    = "\033[33m"
	green     = "\033[32m"
	cyan      = "\033[36m"
	white     = "\033[97m"
	dimWhite  = "\033[37m"
	bgRed     = "\033[41m"
	faintCyan = "\033[2;36m"
)

// Options controls rendering behaviour.
type Options struct {
	NoColor bool
}

type Renderer struct {
	opts Options
	w    io.Writer
}

func New(w io.Writer, opts Options) *Renderer {
	return &Renderer{w: w, opts: opts}
}

func (r *Renderer) color(codes ...string) string {
	if r.opts.NoColor {
		return ""
	}
	return strings.Join(codes, "")
}

func (r *Renderer) print(s string) {
	fmt.Fprint(r.w, s)
}

func (r *Renderer) println(s string) {
	fmt.Fprintln(r.w, s)
}

// RenderResults renders results in pretty format.
// toRender is the list of results to show issues for.
// allResults is used for the overall summary.
func (r *Renderer) RenderResults(toRender []scanner.Result, allResults []scanner.Result) {
	totalErrors := 0
	totalWarns := 0
	totalFiles := len(allResults)
	cleanFiles := 0

	for _, res := range allResults {
		errors, warns := countIssues(res.Issues)
		totalErrors += errors
		totalWarns += warns
		if errors == 0 && warns == 0 {
			cleanFiles++
		}
	}

	// Header
	if totalFiles > 1 {
		r.println(r.color(bold, white) + fmt.Sprintf("yuc: scanning %d files", totalFiles) + r.color(reset))
	} else if totalFiles == 1 {
		r.println(r.color(bold, white) + fmt.Sprintf("yuc: scanning %q", allResults[0].File) + r.color(reset))
	}
	r.println("")

	for _, res := range toRender {
		if totalFiles > 1 && len(res.Issues) > 0 {
			r.println(r.color(bold, white) + "  ┌─ " + res.File + r.color(reset))
		}

		if len(res.Issues) == 0 {
			if totalFiles == 1 {
				r.println(r.color(bold, green) + "  ✓  No issues found." + r.color(reset))
			}
			continue
		}

		for _, issue := range res.Issues {
			r.renderIssue(issue, totalFiles > 1)
		}
		if totalFiles > 1 {
			r.println("")
		}
	}

	// Summary line
	divider := strings.Repeat("─", 62)
	r.println(r.color(dimWhite) + divider + r.color(reset))

	if totalFiles > 1 {
		dirtyFiles := totalFiles - cleanFiles
		r.print(r.color(bold, white) + fmt.Sprintf("  Scanned %d files: ", totalFiles) + r.color(reset))
		r.print(r.color(bold, green) + fmt.Sprintf("%d clean", cleanFiles) + r.color(reset))
		if dirtyFiles > 0 {
			r.print(r.color(dimWhite) + " · " + r.color(reset))
			r.print(r.color(bold, red) + fmt.Sprintf("%d with issues", dirtyFiles) + r.color(reset))
		}
		r.println("")
	}

	r.print("  ")
	if totalErrors > 0 {
		r.print(r.color(bold, red) + fmt.Sprintf("%d error", totalErrors))
		if totalErrors != 1 {
			r.print("s")
		}
		r.print(r.color(reset))
	} else {
		r.print(r.color(bold, green) + "0 errors" + r.color(reset))
	}
	r.print(r.color(dimWhite) + " · " + r.color(reset))
	if totalWarns > 0 {
		r.print(r.color(bold, yellow) + fmt.Sprintf("%d warning", totalWarns))
		if totalWarns != 1 {
			r.print("s")
		}
		r.print(r.color(reset))
	} else {
		r.print(r.color(bold, green) + "0 warnings" + r.color(reset))
	}
	if totalFiles == 1 && len(allResults) > 0 {
		r.print(r.color(dimWhite) + fmt.Sprintf("  in %s", allResults[0].File) + r.color(reset))
	}
	r.println("")
}

func (r *Renderer) renderIssue(issue scanner.Issue, multiFile bool) {
	// Severity badge
	var sevLabel, sevColor string
	switch issue.Severity {
	case "error":
		sevLabel = "ERROR"
		sevColor = r.color(bold, red)
	default:
		sevLabel = "WARN "
		sevColor = r.color(bold, yellow)
	}

	prefix := "  "
	if multiFile {
		prefix = "  │  "
	}

	// Location + category + codepoint
	location := fmt.Sprintf("line %d, col %d", issue.Line, issue.Col)
	cpStr := fmt.Sprintf("U+%04X", issue.Codepoint)

	r.println(fmt.Sprintf("%s%s%s%s  %s%s%s  %s[%s]%s  %s%s%s",
		prefix,
		sevColor, sevLabel, r.color(reset),
		r.color(dimWhite), location, r.color(reset),
		r.color(faintCyan), issue.Category, r.color(reset),
		r.color(bold), cpStr+" "+issue.CharName, r.color(reset),
	))

	// Source line with pointer
	displayLine := sanitiseForDisplay(issue.LineText)
	r.println(fmt.Sprintf("%s          %s%s%s", prefix, r.color(dimWhite), displayLine, r.color(reset)))

	// Pointer arrow
	// Build pointer: count visible runes up to col-1, accounting for tab expansion
	pointerOffset := runeDisplayWidth(issue.LineText, issue.Col-1)
	pointer := strings.Repeat(" ", pointerOffset+10) + r.color(bold, cyan) + "^" + r.color(reset)
	r.println(prefix + pointer)

	// Reason
	r.println(fmt.Sprintf("%s          %s%s%s", prefix, r.color(dimWhite), issue.Reason, r.color(reset)))
	r.println("")
}

// sanitiseForDisplay replaces control and invisible chars with visible stand-ins for the source line preview.
func sanitiseForDisplay(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 || (r >= 0x7F && r <= 0x9F) {
			b.WriteRune('·')
		} else if !utf8.ValidRune(r) {
			b.WriteRune('?')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// runeDisplayWidth returns the number of display columns taken by the first n runes.
func runeDisplayWidth(s string, n int) int {
	count := 0
	i := 0
	for _, _ = range s {
		if i >= n {
			break
		}
		count++
		i++
	}
	return count
}

// ---- Helpers ---------------------------------------------------------------

func countIssues(issues []scanner.Issue) (errors, warns int) {
	for _, i := range issues {
		if i.Severity == "error" {
			errors++
		} else {
			warns++
		}
	}
	return
}

// HasErrors returns true if any result has error-severity issues.
func HasErrors(results []scanner.Result) bool {
	for _, res := range results {
		for _, i := range res.Issues {
			if i.Severity == "error" {
				return true
			}
		}
	}
	return false
}

// HasWarnings returns true if any result has warn-severity issues.
func HasWarnings(results []scanner.Result) bool {
	for _, res := range results {
		for _, i := range res.Issues {
			if i.Severity == "warn" {
				return true
			}
		}
	}
	return false
}
