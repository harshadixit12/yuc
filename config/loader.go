package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// LoadFile reads and merges a user config file into the default config.
// The file format is a simplified YAML that we parse manually (no external deps).
func LoadFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("cannot open config file %q: %w", path, err)
	}
	defer f.Close()

	raw, err := parseConfigYAML(f)
	if err != nil {
		return Config{}, fmt.Errorf("config parse error in %q: %w", path, err)
	}

	return mergeConfig(raw)
}

// mergeConfig takes the parsed raw config and merges it with defaults.
func mergeConfig(raw RawConfig) (Config, error) {
	cfg := DefaultConfig()

	// Apply category overrides
	for catName, override := range raw.Categories {
		cat, ok := cfg.Categories[catName]
		if !ok {
			return cfg, fmt.Errorf("unknown category %q in config; run --list-categories for valid names", catName)
		}
		if override.Severity != "" {
			if override.Severity != SeverityError && override.Severity != SeverityWarn {
				return cfg, fmt.Errorf("invalid severity %q for category %q; must be %q or %q",
					override.Severity, catName, SeverityError, SeverityWarn)
			}
			cat.Severity = override.Severity
		}
		if override.Enabled != nil {
			cat.Enabled = *override.Enabled
		}
		cfg.Categories[catName] = cat
	}

	// Rebuild char list with updated category severities, filtering disabled categories
	filtered := cfg.Chars[:0:len(cfg.Chars)]
	filtered = filtered[:0]
	for _, ch := range cfg.Chars {
		cat := cfg.Categories[ch.Category]
		if !cat.Enabled {
			continue
		}
		ch.Severity = cat.Severity // category severity wins
		filtered = append(filtered, ch)
	}
	cfg.Chars = filtered

	// Add custom chars
	for i, cc := range raw.CustomChars {
		cp, err := parseCodepoint(cc.Codepoint)
		if err != nil {
			return cfg, fmt.Errorf("custom_chars[%d]: invalid codepoint %q: %w", i, cc.Codepoint, err)
		}
		sev := cc.Severity
		if sev == "" {
			sev = SeverityWarn
		}
		if sev != SeverityError && sev != SeverityWarn {
			return cfg, fmt.Errorf("custom_chars[%d]: invalid severity %q", i, sev)
		}
		cfg.Chars = append(cfg.Chars, CharEntry{
			Codepoint: cp,
			Name:      cc.Name,
			Category:  "CUSTOM",
			Severity:  sev,
			Reason:    cc.Reason,
		})
	}

	// Build allowlist
	for _, cpStr := range raw.Allowlist {
		cp, err := parseCodepoint(cpStr)
		if err != nil {
			return cfg, fmt.Errorf("allowlist: invalid codepoint %q: %w", cpStr, err)
		}
		cfg.Allowlist[cp] = true
	}

	return cfg, nil
}

// parseCodepoint parses "U+202E" or "0x202E" or "202E" into a rune.
func parseCodepoint(s string) (rune, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "U+")
	s = strings.TrimPrefix(s, "u+")
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	n, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		return 0, err
	}
	return rune(n), nil
}

// parseConfigYAML is a minimal hand-rolled parser for our config format.
// We only support the subset of YAML used by yuc config files:
// - top-level keys: categories, custom_chars, allowlist
// - categories is a map of string → {severity, enabled}
// - custom_chars is a sequence of {codepoint, name, severity, reason}
// - allowlist is a sequence of strings
func parseConfigYAML(r io.Reader) (RawConfig, error) {
	var raw RawConfig
	raw.Categories = make(map[string]RawCategory)

	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return raw, err
	}

	i := 0
	for i < len(lines) {
		line := lines[i]
		stripped := strings.TrimSpace(line)

		// Skip blanks and comments
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			i++
			continue
		}

		// Top-level keys (no leading spaces)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			key := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(line, ":", 2)[0]), "")
			switch key {
			case "categories":
				i++
				i = parseCategoriesBlock(lines, i, &raw)
			case "custom_chars":
				i++
				i = parseCustomCharsBlock(lines, i, &raw)
			case "allowlist":
				i++
				i = parseAllowlistBlock(lines, i, &raw)
			default:
				i++
			}
		} else {
			i++
		}
	}
	return raw, nil
}

func parseCategoriesBlock(lines []string, i int, raw *RawConfig) int {
	for i < len(lines) {
		line := lines[i]
		stripped := strings.TrimSpace(line)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			i++
			continue
		}
		indent := leadingSpaces(line)
		if indent == 0 {
			break // back to top level
		}
		// Category name line: "  BIDI_CONTROL:"
		if !strings.HasPrefix(stripped, "-") && strings.HasSuffix(strings.TrimSpace(stripped), ":") {
			catName := strings.TrimSuffix(strings.TrimSpace(stripped), ":")
			var rc RawCategory
			i++
			for i < len(lines) {
				inner := lines[i]
				innerStripped := strings.TrimSpace(inner)
				if innerStripped == "" || strings.HasPrefix(innerStripped, "#") {
					i++
					continue
				}
				if leadingSpaces(inner) <= indent {
					break
				}
				kv := strings.SplitN(innerStripped, ":", 2)
				if len(kv) == 2 {
					k := strings.TrimSpace(kv[0])
					v := strings.TrimSpace(kv[1])
					v = stripInlineComment(v)
					switch k {
					case "severity":
						rc.Severity = v
					case "enabled":
						b := v == "true"
						rc.Enabled = &b
					}
				}
				i++
			}
			raw.Categories[catName] = rc
		} else {
			i++
		}
	}
	return i
}

func parseCustomCharsBlock(lines []string, i int, raw *RawConfig) int {
	for i < len(lines) {
		line := lines[i]
		stripped := strings.TrimSpace(line)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			i++
			continue
		}
		if leadingSpaces(line) == 0 {
			break
		}
		if strings.HasPrefix(stripped, "-") {
			var cc RawCustomChar
			// Handle inline key-value on the dash line: "  - codepoint: U+FFFD"
			inlineContent := strings.TrimSpace(strings.TrimPrefix(stripped, "-"))
			if strings.Contains(inlineContent, ":") {
				applyCustomCharKV(&cc, inlineContent)
			}
			i++
			for i < len(lines) {
				inner := lines[i]
				innerStripped := strings.TrimSpace(inner)
				if innerStripped == "" || strings.HasPrefix(innerStripped, "#") {
					i++
					continue
				}
				if strings.HasPrefix(innerStripped, "-") || leadingSpaces(inner) == 0 {
					break
				}
				applyCustomCharKV(&cc, innerStripped)
				i++
			}
			raw.CustomChars = append(raw.CustomChars, cc)
		} else {
			i++
		}
	}
	return i
}

func applyCustomCharKV(cc *RawCustomChar, line string) {
	kv := strings.SplitN(line, ":", 2)
	if len(kv) != 2 {
		return
	}
	k := strings.TrimSpace(kv[0])
	v := strings.TrimSpace(kv[1])
	v = stripInlineComment(v)
	switch k {
	case "codepoint":
		cc.Codepoint = stripQuotes(v)
	case "name":
		cc.Name = stripQuotes(v)
	case "severity":
		cc.Severity = v
	case "reason":
		cc.Reason = stripQuotes(v)
	}
}

func parseAllowlistBlock(lines []string, i int, raw *RawConfig) int {
	for i < len(lines) {
		line := lines[i]
		stripped := strings.TrimSpace(line)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			i++
			continue
		}
		if leadingSpaces(line) == 0 {
			break
		}
		if strings.HasPrefix(stripped, "-") {
			val := strings.TrimSpace(strings.TrimPrefix(stripped, "-"))
			val = stripInlineComment(val)
			val = stripQuotes(val)
			if val != "" {
				raw.Allowlist = append(raw.Allowlist, val)
			}
		}
		i++
	}
	return i
}

func leadingSpaces(s string) int {
	count := 0
	for _, c := range s {
		if c == ' ' {
			count++
		} else if c == '\t' {
			count += 2
		} else {
			break
		}
	}
	return count
}

func stripInlineComment(s string) string {
	// Remove trailing # comments (only if preceded by whitespace)
	if idx := strings.Index(s, " #"); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}
	return strings.TrimSpace(s)
}

func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
