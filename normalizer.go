package ctcalgo

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

var reRemoveBrackets = regexp.MustCompile(`[\(\)\[\]\{\}＜＞<>]`)

func removeBrackets(s string) string {
	return reRemoveBrackets.ReplaceAllString(s, " ")
}

var reNormalizeAllSpaces = regexp.MustCompile(`\s+`)

func normalizeAllSpaces(s string) string {
	return reNormalizeAllSpaces.ReplaceAllString(s, " ")
}

func isDigitToken(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func removeDigitTokens(text string) string {
	parts := strings.Fields(text)
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		if isDigitToken(p) {
			continue
		}
		out = append(out, p)
	}

	return strings.Join(out, " ")
}

var rmSpecialPattern = regexp.MustCompile(
	"[\u200e\u200c\u0656-\u0657\u200b\u064b-\u0652\u202c\u200f\u202a]",
)

func removeSpecial(text string) string {
	return rmSpecialPattern.ReplaceAllString(text, "")
}

var mappingRules = []struct {
	pattern *regexp.Regexp
	replace string
}{
	{regexp.MustCompile(`&lt;`), ""},
	{regexp.MustCompile(`&gt;`), ""},
	{regexp.MustCompile(`&nbsp`), ""},
	{regexp.MustCompile(`(\S+)[` + "\u201b\u2019\u2018" + `](\S+)`), "${1}'${2}"},
}

func applyMappings(text string) string {
	for _, rule := range mappingRules {
		text = rule.pattern.ReplaceAllString(text, rule.replace)
	}
	return text
}

var replaceBracketsWithSpacePattern = regexp.MustCompile(`\([^\)]*\)`)

func replaceBracketsWithSpace(text string) string {
	return replaceBracketsWithSpacePattern.ReplaceAllString(text, " ")
}

func normalizeText(text string, lowerCase bool, removeNumbers bool, removeBrackets bool) string {

	text = norm.NFKC.String(text)

	// text = unidecode.Unidecode(text)

	if lowerCase {
		c := cases.Lower(language.Und)
		text = c.String(text)
	}

	if removeBrackets {
		text = replaceBracketsWithSpace(text)
	}

	text = applyMappings(text)

	text = replacePunctWithSpace(text)

	text = removeSpecial(text)

	if removeNumbers {
		text = removeDigitTokens(text)
	}

	text = normalizeAllSpaces(text)

	return strings.TrimSpace(text)
}
