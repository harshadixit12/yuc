package config

// Severity levels
const (
	SeverityError = "error"
	SeverityWarn  = "warn"
)

// CharEntry describes a single dangerous codepoint.
type CharEntry struct {
	Codepoint rune
	Name      string
	Category  string
	Severity  string
	Reason    string
}

// Category holds metadata about a risk category.
type Category struct {
	Name     string
	Severity string
	Enabled  bool
}

// Config is the merged, resolved configuration used by the scanner.
type Config struct {
	Categories map[string]Category
	Chars      []CharEntry // all active chars (built-in + custom), after applying overrides
	Allowlist  map[rune]bool
}

// RawConfig is the deserialized user config file.
type RawConfig struct {
	Categories  map[string]RawCategory `yaml:"categories"`
	CustomChars []RawCustomChar        `yaml:"custom_chars"`
	Allowlist   []string               `yaml:"allowlist"`
}

type RawCategory struct {
	Severity string `yaml:"severity"`
	Enabled  *bool  `yaml:"enabled"`
}

type RawCustomChar struct {
	Codepoint string `yaml:"codepoint"`
	Name      string `yaml:"name"`
	Severity  string `yaml:"severity"`
	Reason    string `yaml:"reason"`
}

// defaultCategories defines the built-in category metadata.
var defaultCategories = map[string]Category{
	"BIDI_CONTROL":       {Name: "BIDI_CONTROL", Severity: SeverityError, Enabled: true},
	"ZERO_WIDTH":         {Name: "ZERO_WIDTH", Severity: SeverityWarn, Enabled: true},
	"HOMOGLYPH":          {Name: "HOMOGLYPH", Severity: SeverityWarn, Enabled: true},
	"LINE_BREAK_CONTROL": {Name: "LINE_BREAK_CONTROL", Severity: SeverityError, Enabled: true},
	"TAG_CONFUSION":      {Name: "TAG_CONFUSION", Severity: SeverityError, Enabled: true},
	"OVERLONG_SPACE":     {Name: "OVERLONG_SPACE", Severity: SeverityWarn, Enabled: true},
}

// defaultChars is the built-in list of dangerous codepoints.
var defaultChars = []CharEntry{
	// BIDI_CONTROL
	{0x202A, "LEFT-TO-RIGHT EMBEDDING", "BIDI_CONTROL", SeverityError,
		"BiDi embedding can visually reorder displayed text, masking the true content (Trojan Source attack vector)"},
	{0x202B, "RIGHT-TO-LEFT EMBEDDING", "BIDI_CONTROL", SeverityError,
		"BiDi embedding can visually reorder displayed text, masking the true content (Trojan Source attack vector)"},
	{0x202C, "POP DIRECTIONAL FORMATTING", "BIDI_CONTROL", SeverityError,
		"Closes a BiDi embedding scope; paired with other BiDi chars to disguise content"},
	{0x202D, "LEFT-TO-RIGHT OVERRIDE", "BIDI_CONTROL", SeverityError,
		"Forcibly overrides text direction, making malicious values appear benign in editors and diffs"},
	{0x202E, "RIGHT-TO-LEFT OVERRIDE", "BIDI_CONTROL", SeverityError,
		"Forcibly overrides text direction, making malicious values appear benign in editors and diffs"},
	{0x2066, "LEFT-TO-RIGHT ISOLATE", "BIDI_CONTROL", SeverityError,
		"BiDi isolate can disguise the visual order of config values"},
	{0x2067, "RIGHT-TO-LEFT ISOLATE", "BIDI_CONTROL", SeverityError,
		"BiDi isolate can disguise the visual order of config values"},
	{0x2068, "FIRST STRONG ISOLATE", "BIDI_CONTROL", SeverityError,
		"BiDi isolate can disguise the visual order of config values"},
	{0x2069, "POP DIRECTIONAL ISOLATE", "BIDI_CONTROL", SeverityError,
		"Closes a BiDi isolate scope; paired with other BiDi chars to disguise content"},
	{0x200F, "RIGHT-TO-LEFT MARK", "BIDI_CONTROL", SeverityError,
		"Invisible mark that changes text directionality; can corrupt key names and values"},

	// ZERO_WIDTH
	{0x200B, "ZERO WIDTH SPACE", "ZERO_WIDTH", SeverityWarn,
		"Invisible character; can silently corrupt key names, causing lookups to fail unexpectedly"},
	{0x200C, "ZERO WIDTH NON-JOINER", "ZERO_WIDTH", SeverityWarn,
		"Invisible character; may appear in copy-pasted content and corrupt string comparisons"},
	{0x200D, "ZERO WIDTH JOINER", "ZERO_WIDTH", SeverityWarn,
		"Invisible character used to join emoji sequences; unexpected in YAML config files"},
	{0xFEFF, "ZERO WIDTH NO-BREAK SPACE (BOM)", "ZERO_WIDTH", SeverityWarn,
		"UTF-8 BOM; some parsers choke on mid-file BOMs and it creates invisible key-name prefixes"},
	{0x2060, "WORD JOINER", "ZERO_WIDTH", SeverityWarn,
		"Invisible formatting character; no business being in a YAML config value"},
	{0x180E, "MONGOLIAN VOWEL SEPARATOR", "ZERO_WIDTH", SeverityWarn,
		"Zero-width character that can silently corrupt string values and map keys"},

	// HOMOGLYPH — Cyrillic lookalikes
	{0x0430, "CYRILLIC SMALL LETTER A", "HOMOGLYPH", SeverityWarn,
		"Visually identical to ASCII 'a' (U+0061); may deceive readers or cause unexpected key mismatches"},
	{0x0435, "CYRILLIC SMALL LETTER IE", "HOMOGLYPH", SeverityWarn,
		"Visually identical to ASCII 'e' (U+0065); can silently break key name lookups"},
	{0x043E, "CYRILLIC SMALL LETTER O", "HOMOGLYPH", SeverityWarn,
		"Visually identical to ASCII 'o' (U+006F); may deceive readers or break comparisons"},
	{0x0440, "CYRILLIC SMALL LETTER ER", "HOMOGLYPH", SeverityWarn,
		"Visually identical to ASCII 'r' (U+0072); can corrupt identifiers and key names"},
	{0x0441, "CYRILLIC SMALL LETTER ES", "HOMOGLYPH", SeverityWarn,
		"Visually identical to ASCII 'c' (U+0063); may break config key lookups"},
	{0x0445, "CYRILLIC SMALL LETTER HA", "HOMOGLYPH", SeverityWarn,
		"Visually identical to ASCII 'x' (U+0078); can silently introduce unexpected key names"},
	{0x0443, "CYRILLIC SMALL LETTER U", "HOMOGLYPH", SeverityWarn,
		"Visually similar to ASCII 'y' (U+0079) or 'u'; can deceive readers of config values"},
	{0x0455, "CYRILLIC SMALL LETTER DZE", "HOMOGLYPH", SeverityWarn,
		"Visually identical to ASCII 's' (U+0073); can corrupt key or value strings"},
	// Greek lookalikes
	{0x03BF, "GREEK SMALL LETTER OMICRON", "HOMOGLYPH", SeverityWarn,
		"Visually identical to ASCII 'o' (U+006F); can deceive readers and break key matching"},
	{0x03BD, "GREEK SMALL LETTER NU", "HOMOGLYPH", SeverityWarn,
		"Visually similar to ASCII 'v' (U+0076); can silently corrupt identifiers"},
	{0x03C5, "GREEK SMALL LETTER UPSILON", "HOMOGLYPH", SeverityWarn,
		"Visually similar to ASCII 'u' or 'v'; may confuse readers or tools"},
	// Latin extended lookalikes
	{0x0251, "LATIN SMALL LETTER ALPHA", "HOMOGLYPH", SeverityWarn,
		"Visually similar to ASCII 'a' (U+0061); unexpected in YAML identifiers"},
	{0x0261, "LATIN SMALL LETTER SCRIPT G", "HOMOGLYPH", SeverityWarn,
		"Visually similar to ASCII 'g' (U+0067); can corrupt key names subtly"},

	// LINE_BREAK_CONTROL
	{0x0085, "NEXT LINE (NEL)", "LINE_BREAK_CONTROL", SeverityError,
		"Non-standard line ending; some YAML parsers treat this as a line break, others don't — causes inconsistent parsing"},
	{0x2028, "LINE SEPARATOR", "LINE_BREAK_CONTROL", SeverityError,
		"Unicode line separator; can split string values unexpectedly in some YAML parsers"},
	{0x2029, "PARAGRAPH SEPARATOR", "LINE_BREAK_CONTROL", SeverityError,
		"Unicode paragraph separator; can introduce phantom newlines in multi-line string values"},

	// TAG_CONFUSION
	{0xFE56, "SMALL QUESTION MARK", "TAG_CONFUSION", SeverityError,
		"Presentation form of '?'; in some parsers triggers YAML tag parsing, breaking document structure"},
	{0xFE15, "PRESENTATION FORM FOR VERTICAL EXCLAMATION MARK", "TAG_CONFUSION", SeverityError,
		"Presentation form of '!'; can trigger YAML tag or directive parsing in lenient parsers"},

	// OVERLONG_SPACE
	{0x00A0, "NO-BREAK SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Looks like a regular space but isn't; can corrupt YAML indentation in indentation-sensitive blocks"},
	{0x2000, "EN QUAD", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; breaks indentation-sensitive YAML parsing"},
	{0x2001, "EM QUAD", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; breaks indentation-sensitive YAML parsing"},
	{0x2002, "EN SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; can invisibly corrupt YAML indentation"},
	{0x2003, "EM SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; can invisibly corrupt YAML indentation"},
	{0x2004, "THREE-PER-EM SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; breaks YAML indentation"},
	{0x2005, "FOUR-PER-EM SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; breaks YAML indentation"},
	{0x2006, "SIX-PER-EM SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; breaks YAML indentation"},
	{0x2007, "FIGURE SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; can silently corrupt config structure"},
	{0x2008, "PUNCTUATION SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; not valid YAML indentation"},
	{0x2009, "THIN SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space character; can silently corrupt YAML indentation"},
	{0x200A, "HAIR SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Near-invisible space character; can corrupt key names and indentation"},
	{0x3000, "IDEOGRAPHIC SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Full-width space; visually indistinguishable from two regular spaces, corrupts indentation"},
	{0x202F, "NARROW NO-BREAK SPACE", "OVERLONG_SPACE", SeverityWarn,
		"Non-standard space; visually similar to regular space but breaks YAML indentation rules"},
}

// DefaultConfig returns the default configuration with all built-in rules active.
func DefaultConfig() Config {
	cats := make(map[string]Category, len(defaultCategories))
	for k, v := range defaultCategories {
		cats[k] = v
	}
	return Config{
		Categories: cats,
		Chars:      append([]CharEntry(nil), defaultChars...),
		Allowlist:  make(map[rune]bool),
	}
}

// AllCategories returns the default category list for --list-categories.
func AllCategories() map[string]Category {
	result := make(map[string]Category, len(defaultCategories))
	for k, v := range defaultCategories {
		result[k] = v
	}
	return result
}
