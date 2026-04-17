package ctcalgo

import (
	"strings"
	"testing"
)

func TestPreprocessText(t *testing.T) {

	lines, err := readFileLines("assets/preprocess_test.lst")
	if err != nil {
		panic(err)
	}

	for _, l := range lines {

		parts := strings.Split(l, "\t")
		if len(parts) != 3 {
			continue
		}

		original := parts[0]
		starredTokensTarget := parts[1]
		starredTextTarget := parts[2]

		starredTokens, starredText := preprocessText(strings.TrimSpace(original), true)

		starredTokensHyp := strings.Join(starredTokens, "  ")
		starredTextHyp := strings.Join(starredText, " ")

		if starredTokensHyp != starredTokensTarget {
			t.Errorf("❌ FAIL: original:%s|expected:%s|got:%s|", original, starredTokensTarget, starredTokensHyp)
			continue
		}

		if starredTextHyp != starredTextTarget {
			t.Errorf("❌ FAIL: original:%s|expected:%s|got:%s|", original, starredTextTarget, starredTextHyp)
			continue
		}
	}
}
