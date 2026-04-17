package ctcalgo

import (
	"strings"
	"testing"
)

func TestTextNormalizeAndRomanize(t *testing.T) {

	lines, err := readFileLines("assets/norm_romn_test.lst")
	if err != nil {
		panic(err)
	}

	sep := " "

	for _, l := range lines {
		parts := strings.Split(l, "\t")
		if len(parts) != 3 {
			continue
		}
		original := parts[0]
		normTarget := parts[1]
		romnTarget := parts[2]

		originalTokens := strings.Split(original, sep)
		normalizedTokens := make([]string, len(originalTokens))
		for i := range originalTokens {
			normalizedTokens[i] = normalizeText(strings.TrimSpace(originalTokens[i]), true, true, false)
		}
		normHyp := strings.Join(normalizedTokens, sep)

		romanizedTokens := getRomanizedTokens(normalizedTokens, "")
		romnHyp := strings.Join(romanizedTokens, "  ") // here i use double space in ref file

		if normHyp != normTarget {
			t.Errorf("❌ FAIL (normalizeText): original:%s|expected:%s|got:%s|", original, normTarget, normHyp)
			continue
		}

		if romnHyp != romnTarget {
			t.Errorf("❌ FAIL (romanizeText): original:%s|expected:%s|got:%s|", original, romnTarget, romnHyp)
			t.Logf("tokens:")
			for _, tk := range romanizedTokens {
				t.Logf("%s", tk)
			}
			continue
		}
	}
}
