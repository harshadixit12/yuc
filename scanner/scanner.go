package scanner

import (
	"bufio"
	"io"

	"github.com/harshadixit12/yuc/config"
)

// Issue represents a single detected problem.
type Issue struct {
	File      string
	Line      int
	Col       int // 1-based rune index within the line
	Category  string
	Severity  string
	Codepoint rune
	CharName  string
	Reason    string
	LineText  string // raw line content for display
}

// Result holds all issues found in a single file/stream.
type Result struct {
	File   string
	Issues []Issue
	Lines  int
}

// Scan reads from r and returns all detected issues.
// filename is used for display and internal reporting.
func Scan(r io.Reader, cfg config.Config, filename string) (Result, error) {
	// Build a lookup map: codepoint → CharEntry (last one wins for duplicates)
	lookup := make(map[rune]config.CharEntry, len(cfg.Chars))
	for _, ch := range cfg.Chars {
		if !cfg.Allowlist[ch.Codepoint] {
			lookup[ch.Codepoint] = ch
		}
	}

	result := Result{File: filename}

	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		result.Lines = lineNum

		colNum := 0
		for _, r := range text {
			colNum++
			if entry, bad := lookup[r]; bad {
				result.Issues = append(result.Issues, Issue{
					File:      filename,
					Line:      lineNum,
					Col:       colNum,
					Category:  entry.Category,
					Severity:  entry.Severity,
					Codepoint: r,
					CharName:  entry.Name,
					Reason:    entry.Reason,
					LineText:  text,
				})
			}
		}
	}

	return result, scanner.Err()
}
