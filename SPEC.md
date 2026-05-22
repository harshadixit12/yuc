# `yuc` — Unicode Hazard Scanner for YAML Files

> _Because the scariest bugs hide in plain sight._

---

## Overview

`yuc` scans YAML files for dangerous, deceptive, or ambiguous Unicode characters that can cause security vulnerabilities, parsing inconsistencies, or silent data corruption. It catches the characters that look innocent but aren't.

---

## Name

**`yuc`** — a yaml unicode check. Short, memorable, and accurate: this tool is meant to scare you about what's lurking in your configs.

---

## Synopsis

```
yuc [flags] <file> [file...]
```

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | `-c` | `string` | `""` | Path to a `.yuc.yaml` config file. Overrides default rules. |
| `--no-color` | | `bool` | `false` | Disable ANSI color output (auto-disabled when not a TTY). |
| `--show-safe` | | `bool` | `false` | Also print lines that passed (for audit trails). |
| `--list-categories` | | `bool` | `false` | Print all known risk categories and exit. |
| `--version` | `-v` | `bool` | `false` | Print version and exit. |
| `--help` | `-h` | `bool` | `false` | Print help and exit. |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Clean — no issues found |
| `1` | Issues found (errors or warnings) |
| `2` | Bad usage / invalid flags / unreadable file |
| `3` | Config file parse error |

---

## Output Format

```
yuc: scanning "deploy.yaml"

  ERROR   line 12, col 8   [BIDI_CONTROL]   U+202E RIGHT-TO-LEFT OVERRIDE
          "  key: val\u202eue"
                   ^ right-to-left override can reverse displayed text, masking true content

  WARN    line 34, col 21  [HOMOGLYPH]      U+0430 CYRILLIC SMALL LETTER A
          "  admin: fаlse"
                       ^ visually identical to ASCII 'a' (U+0061); may deceive readers or parsers

  WARN    line 41, col 1   [ZERO_WIDTH]     U+200B ZERO WIDTH SPACE
          "​key: value"
           ^ invisible character; may silently corrupt key names

──────────────────────────────────────────────────────────
  2 errors · 2 warnings  in deploy.yaml
```

Colors:
- 🔴 `ERROR` — bold red
- 🟡 `WARN` — bold yellow
- ✅ `OK` — bold green (only with `--show-safe`)
- Offending character position marker `^` — cyan
- File name and counts — bold white

---

## Default Risk Categories

These are enabled by default. Each has a **severity** (`error` or `warn`) and a set of codepoints or Unicode ranges.

### `BIDI_CONTROL` — severity: `error`
Bidirectional text control characters. Used in [Trojan Source](https://trojansource.codes/) attacks to visually misrepresent code or config values.

| Codepoint | Name |
|-----------|------|
| U+202A | LEFT-TO-RIGHT EMBEDDING |
| U+202B | RIGHT-TO-LEFT EMBEDDING |
| U+202C | POP DIRECTIONAL FORMATTING |
| U+202D | LEFT-TO-RIGHT OVERRIDE |
| U+202E | RIGHT-TO-LEFT OVERRIDE |
| U+2066 | LEFT-TO-RIGHT ISOLATE |
| U+2067 | RIGHT-TO-LEFT ISOLATE |
| U+2068 | FIRST STRONG ISOLATE |
| U+2069 | POP DIRECTIONAL ISOLATE |
| U+200F | RIGHT-TO-LEFT MARK |

### `ZERO_WIDTH` — severity: `warn`
Zero-width and invisible characters. Can silently corrupt key names, string comparisons, and secrets.

| Codepoint | Name |
|-----------|------|
| U+200B | ZERO WIDTH SPACE |
| U+200C | ZERO WIDTH NON-JOINER |
| U+200D | ZERO WIDTH JOINER |
| U+FEFF | ZERO WIDTH NO-BREAK SPACE (BOM) |
| U+2060 | WORD JOINER |
| U+180E | MONGOLIAN VOWEL SEPARATOR |

### `HOMOGLYPH` — severity: `warn`
Characters that look identical (or nearly so) to ASCII letters but are different codepoints. Used in social engineering and typosquatting.

Covers: Cyrillic lookalikes (е, а, о, р, с, х, у), Greek lookalikes (ο, ν, υ), Latin Extended lookalikes (ɑ, ɡ), and others. Full list in source.

### `LINE_BREAK_CONTROL` — severity: `error`
Non-standard line ending characters that can confuse YAML parsers and split values unexpectedly.

| Codepoint | Name |
|-----------|------|
| U+0085 | NEXT LINE (NEL) |
| U+2028 | LINE SEPARATOR |
| U+2029 | PARAGRAPH SEPARATOR |

### `TAG_CONFUSION` — severity: `error`
Characters that are valid YAML tag characters in some parsers but not others, leading to inconsistent parsing.

| Codepoint | Name |
|-----------|------|
| U+FE56 | SMALL QUESTION MARK |
| U+FE15 | PRESENTATION FORM FOR VERTICAL EXCLAMATION MARK |

### `OVERLONG_SPACE` — severity: `warn`
Space-like characters that are not regular spaces (U+0020). These can mislead indentation-sensitive parsers.

| Codepoint | Name |
|-----------|------|
| U+00A0 | NO-BREAK SPACE |
| U+2000–U+200A | EN QUAD through HAIR SPACE |
| U+3000 | IDEOGRAPHIC SPACE |
| U+202F | NARROW NO-BREAK SPACE |

---

## Config File Format (`.yuc.yaml`)

Users can place a config file anywhere and pass it via `--config`. yuc does **not** auto-discover config files.

```yaml
# .yuc.yaml

# Override severity for a whole category, or disable it entirely
categories:
  HOMOGLYPH:
    severity: error       # bump from warn → error
  OVERLONG_SPACE:
    enabled: false        # disable entirely
  ZERO_WIDTH:
    severity: warn        # keep as warn (default, explicit)

# Add custom codepoints to watch for
custom_chars:
  - codepoint: "U+2116"   # № NUMERO SIGN
    name: "NUMERO SIGN"
    severity: warn
    reason: "Numero sign can be confused with 'No' in config values; use plain text instead"

  - codepoint: "U+FFFD"   # REPLACEMENT CHARACTER
    name: "REPLACEMENT CHARACTER"
    severity: error
    reason: "Replacement character indicates broken/lossy encoding — this file has encoding damage"

# Characters to explicitly allowlist (suppress all warnings for these)
allowlist:
  - "U+200C"  # We use zero-width non-joiner intentionally in i18n keys
```

Config merges with defaults: categories not mentioned inherit their defaults. `custom_chars` are additive. `allowlist` suppresses specific codepoints globally.

---

## Multiple File Support

```bash
yuc config/*.yaml
yuc **/*.yaml   # shell glob expansion
```

When scanning multiple files, each file gets its own section. Summary at end:

```
══════════════════════════════════════════════════════════════
  Scanned 7 files:  3 clean · 2 with warnings · 2 with errors
```

Exit code `1` if any file has issues.

---

## Examples

```bash
# Basic scan
yuc config.yaml

# Use custom config
yuc --config .yuc.yaml values.yaml

# Audit multiple files
yuc --show-safe config/*.yaml

# List all built-in categories
yuc --list-categories
```

---

## Security Rationale

YAML is used everywhere: Kubernetes manifests, CI pipelines, Helm charts, Ansible playbooks, app configs. It is a common vector for:

- **Trojan Source**: BiDi characters make malicious values appear benign in editors/diffs
- **Key spoofing**: Zero-width chars create keys that look identical but aren't (`"admin"` vs `"adm​in"`)
- **Encoding corruption**: Replacement chars or NEL breaks indicate damaged files being silently accepted
- **Homoglyph confusion**: Configs edited by non-ASCII keyboard users may accidentally introduce lookalikes

`yuc` is a fast, composable tool that fits into any pipeline and fails loudly when these issues are present.

---

## Non-Goals

- `yuc` does **not** validate YAML structure or schema (use `yamllint` for that)
- It does **not** parse the YAML AST — it works at the raw byte/rune level, line by line
- It does **not** auto-fix files (no `--fix` flag)

---

## Implementation Notes (Go)

- Pure Go, zero runtime dependencies outside stdlib + `gopkg.in/yaml.v3` (for config file parsing only)
- Scanning is rune-level, not YAML-AST-level — intentionally, so it catches issues in comments, keys, and values uniformly
- ANSI color via a tiny internal package (no external color lib needed)
- Designed to be embeddable as a library: `yuc/scanner` package exposes `Scan(r io.Reader, cfg Config) ([]Issue, error)`
