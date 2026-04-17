package ctcalgo

import (
	"regexp"
	"strings"

	"github.com/alexsergivan/transliterator"
)

var romanizer = transliterator.NewTransliterator(nil)

func romanize(s string, isoLang string) string {
	return romanizer.Transliterate(s, isoLang)
}

var nonAllowedRe = regexp.MustCompile(`[^a-z' ]`)
var multiSpaceRe = regexp.MustCompile(` +`)

func normalizeRomanized(text string) string {
	text = strings.ToLower(text)
	text = nonAllowedRe.ReplaceAllString(text, " ")
	text = multiSpaceRe.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

var wsRe = regexp.MustCompile(`\s+`)

func getRomanizedTokens(normTranscripts []string, isoLang string) []string {
	uromans := make([]string, 0, len(normTranscripts))

	for _, transcript := range normTranscripts {
		// romanize
		ot := romanize(transcript, isoLang)

		// trim + insert spaces between characters (like " ".join(ot.strip()))
		ot = strings.TrimSpace(ot)

		chars := []rune(ot)
		spaced := make([]string, 0, len(chars))
		for _, r := range chars {
			spaced = append(spaced, string(r))
		}
		ot = strings.Join(spaced, " ")

		// collapse whitespace
		ot = wsRe.ReplaceAllString(ot, " ")
		ot = strings.TrimSpace(ot)

		// normalize
		normalized := normalizeRomanized(ot)

		uromans = append(uromans, normalized)
	}

	return uromans
}
